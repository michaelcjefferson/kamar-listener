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

## Database
As only new results will be dumped each day, set up a database using SQLite to hold all data, and connect PowerBI to it.

## Testing - 10-4-24
- Had to use desktop IP address rather than localhost as address, and had to remember URL tag, eg. 192.168.1.84/kamar-refresh
- When trying Check and Enable, server displayed the error "failed at authCredentials", and also logged "received and processed check request". KAMAR displayed the error "ERROR: No service name returned"
- When trying Check and Enable with incorrect credentials (username:pa55word), server logged "failed at authCredentials" again, but this time didn't display "received and processed...". KAMAR displayed "ERROR: HTTP/1.1 403 Forbidden"
- Got past "failed at authCredentials" by adding a couple more fields to the SMSDirectoryData field of the response. Now, KAMAR showing this error: "ERROR: Invalid Server - invalid/missing support info URL. Please contact the supplier to update."
- KAMAR check now working - cause of areas was missing fields in the server response to KAMAR's check request. Final thing to add was a privacy statement that was more than just "none".

## Testing - 11-4-24
- TODO: Change port, as 443 might be too open by default, and is used for fmtp (required?)
