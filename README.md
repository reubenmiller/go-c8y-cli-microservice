# Introduction

An example Cumulocity IoT microservice written in go (golang) to add simple endpoints to running `go-c8y-cli` commands

The project uses the unofficial github.com/reubenmiller/go-c8y Cumulocity client modules.

# Getting Started

## Dev container

This project includes a VScode dev container definition to make it easier to install of the project requirements and provide a normalized development environment.

The `~/.cumulocity` folder (created by go-c8y-cli) is mounted from your host OS into your dev container so that all of your existing sessions will be available inside the dev container environment.


## Starting the app locally

1. Clone the project

    ```sh
    git clone https://github.com/reubenmiller/go-c8y-cli-microservice.git
    cd go-c8y-cli-microservice
    ```

2. Open the project in VScode

    ```sh
    code .
    ```

3. When prompted by VScode, rebuild/reopen the project in the dev container. Be patient, it needs to build a docker image.

4. Activate/Create a Cumulocity session pointing to the Cumulocity instance you want to develop against

    ```sh
    set-session
    ```

5. Run the init task to create a dummy microservice placeholder and initialize the `.env` file containing the microservice's bootstrap credentials

    ```
    task init
    ```

6. Start the application via debugger (F5)


## Known Issues

* Ctrl-c does not work to kill the application. You will have to manually stop the process by either killing your console, or the process itself.

## Build via task

Build the Cumulocity microservice zip file by executing

```sh
task build:microservice
```

## Build via docker cli

You can build the microservice locally and upload it to the cumulocity after packing it together with the cumulocity manifest into a zip file.
Find more about that in the [documentation](https://cumulocity.com/guides/microservice-sdk/concept/#overview).

```bash
docker build -t go-c8y-cli-microservice --no-cache --platform linux/amd64 .
```
If you are buulding on e.g. an Apple M1 you have to give --platform as a parameter for building a linux/amd64 container. Otherwise the microservice would not work.
Save the container in a .tar via:


```bash
docker save go-c8y-cli-microservice > image.tar
```

You can either zip manually or via cli

```bash
zip go-c8y-cli-microservice cumulocity.json image.tar
```

## Deploy via task

```
task deploy:microservice
```

## Deploy via UI

You can upload the .zip via the UI in the Administration application.

![Preview](http://g.recordit.co/nkhGv2yMdD.gif)

## Testing the endpoints manually

`go-c8y-cli` can be used to check the microservice's endpoints either when running locally or when hosted in Cumulocity IoT.

When the microservice locally, you will need to add `--host http://localhost:8000` to the command.

### Check prometheus endpoint

```sh
c8y api /prometheus --host http://localhost:8000
```
### Upload a file to import 

```sh
# Run a test command
echo -e "1111\n22222" > input.list
c8y api POST "/commands/importevents/async" --host http://localhost:8000 --force --file "./input.list"
```

```sh
# Run a test command
echo -e "1111\n22222" > input.list
c8y api POST "/service/go-c8y-cli-microservice/commands/importevents/sync" --force --file "./input.list"
```

```
sudo docker run -it --env-file=.env go-c8y-cli-microservice:0.0.1 bash
```

## Execute command

Execute a custom command. Note only creating data is supported

```sh
c8y api POST "/service/go-c8y-cli-microservice/commands/execute/sync" --data "command=c8y devices list | c8y events create --text 'Example event' --type 'c8y_CliExample'"
```
