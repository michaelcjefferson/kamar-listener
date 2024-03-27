# WHS KAMAR Refresh

## Setting up your dev environment
Please download Go v1.20 from the following link and install on your computer:
[Go Download Link](https://go.dev/dl/go1.20.14.windows-amd64.msi)

## To build a new .exe binary
Open terminal, navigate to this directory, and use the following command:
`set CGO_ENABLED=0&& set goos=windows&& go build -o ./bin/kamarRefresh.exe ./cmd/api`

## To run the program in dev mode
Open terminal, navigate to this directory, and use the following command:
`go run ./cmd/api`
Use ctrl+c in the terminal to end the process.

## Optional flags for terminal commands
To see a list of available flags that can be used in either of the above terminal commands, use the following command in this directory:
`go run ./cmd/api -help`

## To test endpoints
If the service is running,  in the terminal navigate to the /test directory inside this folder (from the same computer that the application is running on). Use the following command:
`curl -X POST -d @./**filename**.json -H "Content-Type: application/json" --user "**username**:**password**" localhost:443/kamar-refresh`
You may need to use https://localhost... if HTTPS is turned on.
