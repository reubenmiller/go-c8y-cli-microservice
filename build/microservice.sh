#!/bin/bash

WORK_DIR=$(pwd)
IMAGE_NAME=
TAG_NAME="latest"
DEPLOY_ADDRESS=
DEPLOY_TENANT=
DEPLOY_USER=
DEPLOY_PASSWORD=
APPLICATION_NAME=
APPLICATION_ID=

PACK=1
DEPLOY=1
SUBSCRIBE=1
HELP=1


execute () {
	readInput $@
	cd $WORK_DIR
	if [ "$HELP" == "0" ]
	then
		printHelp
		exit
	fi
	if [ "$PACK" == "1" ] && [ "$DEPLOY" == "1" ] && [ "$SUBSCRIBE" == "1" ]
	then
		echo "[INFO] No goal set. Please set pack, deploy or subscribe"
	fi
	if [ "$PACK" == "0" ]
	then 
		echo "[INFO] Start packaging"
		verifyPackPrerequisits
		clearTarget
		buildImage
		exportImage
		zipFile
		echo "[INFO] End packaging"
	fi
	if [ "$DEPLOY" == "0" ]
	then
		echo "[INFO] Start deployment"
		deploy
		echo "[INFO] End deployment"
	fi
	if [ "$SUBSCRIBE" == "0" ]
	then
		echo "[INFO] Start subsciption"
		subscribe
		echo "[INFO] End subsciption"
	fi
	exit 0
}

readInput () {
	echo "[INFO] Read input"
	while [[ $# -gt 0 ]]
	do
	key="$1"
	case $key in
		pack)
		PACK=0
		shift
		;;	
		deploy)
		DEPLOY=0
		shift
		;;
		subscribe)
		SUBSCRIBE=0
		shift
		;;
		help | --help)
		HELP=0
		shift
		;;
		-dir | --directory)
		WORK_DIR=$2
		shift
		shift
		;;
		-n | --name)
		IMAGE_NAME=$2
		shift
		shift
		;;
		-t | --tag)
		TAG_NAME=$2
		shift
		shift
		;;
		-d | --deploy)
		DEPLOY_ADDRESS=$2
		shift
		shift
		;;
		-u | --user)
		DEPLOY_USER=$2
		shift
		shift
		;;
		-p | --password)
		DEPLOY_PASSWORD=$2
		shift
		shift
		;;
		-te | --tenant)
		DEPLOY_TENANT=$2
		shift
		shift
		;;
		-a | --application)
		APPLICATION_NAME=$2
		shift
		shift
		;;
		-id | --applicationId)
		APPLICATION_ID=$2
		shift
		shift
		;;
		*)
		shift
		;;
	esac
	done
	setDefaults
}

setDefaults () {
	ZIP_NAME="$IMAGE_NAME.zip"
	if [ "x$APPLICATION_NAME" == "x" ]
	then 
		APPLICATION_NAME=$IMAGE_NAME
	fi	
}

printHelp () {
	echo
	echo "Following functions are available. You can run specify them in single execution:"
	echo "	pack - prepares deployable zip file. Requires following stucture:"
	echo "		/docker/Dockerfile"
	echo "		/docker/* - all files within the directory will be included in the docker build"
	echo "		/cumulocity.json "
	echo "	deploy - deploys applicaiton to specified address"
	echo "	subscribe - subscribes tenant to specified microservice application"
	echo "	help | --help - prints help"
	echo 
	echo "Following options are available:"
	echo "	-dir | --directory 		# Working directory. Default value'$(pwd)' "
	echo "	-n   | --name 	 		# Docker image name"
	echo "	-t   | --tag			# Docker tag. Default value 'latest'"
	echo "	-d   | --deploy			# Address of the platform the microservice will be uploaded to"	
	echo "	-u   | --user			# Username used for authentication to the platform"
	echo "	-p   | --password 		# Password used for authentication to the platform"
	echo "	-te  | --tenant			# Tenant used"
	echo "	-a   | --application 	# Name upon which the application will be registered on the platform. Default value from --name parameter"
	echo "	-id  | --applicationId	# Applicaiton used for subscription purposes. Required only for solemn subscribe execution"
}

verifyPackPrerequisits () {
	echo "[INFO] Check input"
	result=0
	verifyParamSet "$IMAGE_NAME" "name"

	isPresent $(find "$WORK_DIR" -maxdepth 1 -name "Dockerfile" | wc -l) "[ERROR] Stopped: missing dockerfile in work directory: $WORK_DIR"
	isPresent $(find . -maxdepth 1 -name "cumulocity.json" | wc -l) "[ERROR] Stopped: missing cumulocity.json in work directory: $WORK_DIR"
	# Find the dockerfile, and set context
	DOCKERFILE_FOLDER=$(echo "$(dirname `find "$WORK_DIR" -maxdepth 1 -name "Dockerfile"`)")

	if [ "$result" == "1" ]
	then
		echo "[WARNING] Pack skiped"
		exit 1
	fi
}

isPresent () {
	present=$1
	if [ "$present" != "1" ]
	then
		echo $2
		result=1
	fi
}

clearTarget () {
	echo "[INFO] Clear target files"
	if [ -f "image.tar" ]; then
		rm "image.tar"
	fi

	if [ -f "$ZIP_NAME" ]; then
		rm "$ZIP_NAME"
	fi
}

buildImage () {
	cd "$DOCKERFILE_FOLDER"
	echo "[INFO] Build image $IMAGE_NAME:$TAG_NAME"
	docker build --build-arg HTTP_PROXY=$HTTP_PROXY --build-arg HTTPS_PROXY=$HTTPS_PROXY --build-arg http_proxy=$HTTP_PROXY --build-arg https_proxy=$HTTPS_PROXY -t $IMAGE_NAME:$TAG_NAME .
}

exportImage () {
	echo "[INFO] Export image"
	docker save $IMAGE_NAME:$TAG_NAME > "image.tar"
}

zipFile () {
	echo "[INFO] Zip file $ZIP_NAME"
	echo "[INFO] Working dir [$( pwd )]"
	zip $ZIP_NAME cumulocity.json "image.tar"

	echo "[INFO] Removing image.tar"
	if [ -f image.tar ]; then
		rm -f image.tar
	fi
}

deploy (){
	verifyDeployPrerequisits
	push
}

verifyDeployPrerequisits () {
	result=0
	verifyParamSet "$IMAGE_NAME" "name"
	verifyParamSet "$DEPLOY_ADDRESS" "address"
	verifyParamSet "$DEPLOY_TENANT" "tenant"
	verifyParamSet "$DEPLOY_USER" "user"
	verifyParamSet "$DEPLOY_PASSWORD" "password"
	
	if [ "$result" == "1" ]
	then
		echo "[WARNING] Deployment skiped"
		exit 1
	fi
}

verifyParamSet (){
	if [ "x$1" == "x" ]
	then
		echo "[WARNING] Missing parameter: $2"
		result=1
	fi		
}

push (){
	authorization="Basic $(echo -n "$DEPLOY_USER:$DEPLOY_PASSWORD" | base64)"
	
	APPLICATION_ID=$(getApplicationId)
	if [ "x$APPLICATION_ID" == "xnull" ]
	then
		echo "[INFO] Application with name $APPLICATION_NAME not found, add new application"
		createApplication $authorization
		APPLICATION_ID=$(getApplicationId)
		if [ "x$APPLICATION_ID" == "xnull" ]
		then
			echo "[ERROR] Could not create application"
			exit 1
		fi
	fi
	echo "[INFO] Application name: $APPLICATION_NAME id: $APPLICATION_ID"
	
	uploadFile	
}	

getApplicationId () {
	resp=$(curl -s -H "Authorization: $authorization" "$DEPLOY_ADDRESS/application/applicationsByName/$APPLICATION_NAME")
		if [ "x$(echo $resp | jq -r .error)" != "xnull" ]
		then
			echo "[ERROR] Error while connecting to platform"
			echo $resp
			exit
		fi
	echo $(echo $resp | jq -r .applications[0].id)
}

createApplication () {
	body="{
			\"name\": \"$APPLICATION_NAME\",
			\"type\": \"MICROSERVICE\",
			\"key\": \"$APPLICATION_NAME-microservice-key\"
		}
	"
	resp=$(curl -X POST -s -d "$body" -H "Authorization: $authorization" -H "Content-type: application/json" "$DEPLOY_ADDRESS/application/applications") 
}

uploadFile () {
	echo "[INFO] Upload file $WORK_DIR/$ZIP_NAME"
	resp=$(curl -F "data=@$WORK_DIR/$ZIP_NAME" -H "Authorization: $authorization" "$DEPLOY_ADDRESS/application/applications/$APPLICATION_ID/binaries")
	if [ "x$(echo $resp | jq -r .error)" != "xnull" ] && [ "x$(echo $resp | jq -r .error)" != "x" ]
	then		
		echo "[WARNING] error durning upload"
		echo "$(echo $resp | jq -r .message)"
	fi
	if [ "x$(echo $resp | jq -r .error)" == "x" ]
	then		
		echo "[INFO] File uploaded"
	fi
}

subscribe () {
	verifySubscribePrerequisits
	authorization="Basic $(echo -n "$DEPLOY_USER:$DEPLOY_PASSWORD" | base64)"
	
	echo "[INFO] Tenant $DEPLOY_TENANT subscription to application $APPLICATION_NAME with id $APPLICATION_ID"
	body="{\"application\":{\"id\": \"$APPLICATION_ID\"}}"
	resp=$(curl -X POST -s -d "$body"  -H "Authorization: $authorization" -H "Content-type: application/json" "$DEPLOY_ADDRESS/tenant/tenants/$DEPLOY_TENANT/applications")
	if [ "x$(echo $resp | jq -r .error)" != "xnull" ] && [ "x$(echo $resp | jq -r .error)" != "x" ]
	then		
		echo "[WARNING] error subscribing tenant to application "
		echo "$(echo $resp | jq -r .message)"
	fi
	if [ "x$(echo $resp | jq -r .error)" == "x" ]
	then		
		echo "[INFO] Tenant $DEPLOY_TENANT subscribed to application $APPLICATION_NAME"
	fi
}

verifySubscribePrerequisits () {
	if [ "x$APPLICATION_ID" == "x" ]
	then
		echo "[ERROR] Subscription not possible uknknown applicaitonId"
		exit 1
	fi	
	verifyDeployPrerequisits
}

execute $@


