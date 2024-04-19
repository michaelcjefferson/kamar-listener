# WHS KAMAR Refresh
This repository allows for a binary to be built that when executed, runs a server that listens for connections from KAMAR's directory services. It is able to receive datasets and save them to an SQLite database for consumption by PowerBI or another service.

## Setting up your dev environment
Please download Go v1.20 from the following link and install on your computer:
[Go Download Link](https://go.dev/dl/go1.20.14.windows-amd64.msi)
Please also download SQLite Studio and install:
[SQLite Studio Download Link](https://sqlitestudio.pl/)
Once SQLite Studio is installed, open this directory in Windows Explorer, and open the /db folder. Copy and paste the template-kamar-directory-service.db file, then rename it to kamar-directory-service.db.
Open SQLite Studio, then in the top left click Add Database, and add kamar-directory-service.db. Connect to it by double-clicking on it in the sidebar on the left. Double click on the Results table - you should now see its structure in the main window. Click on the data tab - for now it will be empty, but this is where data from KAMAR will be populated. You will most likely need to click the Refresh button (blue icon) to see new data once an upload has been completed.

## To build a new .exe binary
Open terminal, navigate to this directory, and use the following command:
`set CGO_ENABLED=0&& set goos=windows&& go build -o ./bin/kamarRefresh.exe ./cmd/api`
This file should be copied across to (in this case) Mark's computer, to the C:\\Listener directory. It can then be double-clicked to run - a new terminal will open up, which will log any connections that the server receives.

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
As only new results will be dumped each day, an SQLite database is used to hold all data, which PowerBI can connect to and consume.

## Testing - 10-4-24
- Had to use desktop IP address rather than localhost as address, and had to remember URL tag, eg. 192.168.1.84/kamar-refresh
- When trying Check and Enable, server displayed the error "failed at authCredentials", and also logged "received and processed check request". KAMAR displayed the error "ERROR: No service name returned"
- When trying Check and Enable with incorrect credentials (username:pa55word), server logged "failed at authCredentials" again, but this time didn't display "received and processed...". KAMAR displayed "ERROR: HTTP/1.1 403 Forbidden"
- Got past "failed at authCredentials" by adding a couple more fields to the SMSDirectoryData field of the response. Now, KAMAR showing this error: "ERROR: Invalid Server - invalid/missing support info URL. Please contact the supplier to update."
- KAMAR check now working - cause of areas was missing fields in the server response to KAMAR's check request. Final thing to add was a privacy statement that was more than just "none".

## Testing - 11-4-24
- TODO: Change port, as 443 might be too open by default, and is used for fmtp (required?)
