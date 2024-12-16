# WHS KAMAR Refresh
This repository allows for a binary to be built that when executed, runs a server that listens for connections from KAMAR's directory services. It is able to receive datasets and save them to an SQLite database for consumption by PowerBI or another service.

## Day-to-day Operation
### Setting things up
1. Build a new .exe file, by following the steps in **To build a new .exe binary**.
2. Copy the kamarRefresh.exe across to C:\\KamarListener\ on whichever computer will be running the listener service. Please also download SQLite Studio and install:
[SQLite Studio Download Link](https://sqlitestudio.pl/). You can open kamar-directory-service.db with this program and view the data it contains.
3. Create a folder in C:\\KamarListener\ called "tls". Copy key.pem and cert.pem files from the development folder into the tls folder you created.
4. Double-click kamarRefresh.exe to begin. It will automatically create a database file called kamar-directory-service.db in the same folder, and set up a results table inside it. It will also open a terminal window which will log any requests that it receives. kamarRefresh.exe is a web server - it will receive HTTP requests from KAMAR, process them, and then push them to the database.
5. Open KAMAR, and go to Setup --> Server --> Directory Services. Fill out the details there (refer to [this page](https://directoryservices.kamar.nz/?listening-service) for help), then click Check and Run, then tick required fields (eg. Results, setting it up with the timeframe required), then click Update to start the preliminary upload. You should see logs appearing in the terminal window running kamarRefresh.exe, as well as rows of results starting to appear in SQLite Studio (need to click the blue refresh button on the Data tab).
6. To connect PowerBI to this data, first download and install the [SQLite ODBC Driver](http://ch-werner.de/sqliteodbc/). Then, follow the instructions in [this video](https://www.youtube.com/watch?v=n5ELoULhQIo).

### If there was an issue
1. If the .exe process stopped (or you need to reset the computer or something), just double-click it to start again - it will reconnect to the same database without overwriting anything.
2. If the problem persists, check IP addresses - kamarRefresh.exe is currently configured to only accept requests from either localhost or IP addresses beginning with 10.100. If this needs to be changed, you can do so on line 139 of middleware.go.

### Starting over
1. To start over with a fresh database, go to C:\\KamarListener\, and delete **both** kamarRefresh.exe **and** kamar-directory-service.db.
2. Follow the steps in **To build a new .exe binary**.
3. Copy kamarRefresh.exe over to C:\\KamarListener\ (ensure that you still have a tls folder here with key and cert inside), and run it - it will create a fresh database for you.

## Development
### Setting up your dev environment
1. Download a copy of this repository to your computer (or clone it).
2. Please download Go v1.20 (or a later version) from the following link and install on your computer:
[Go Download Link](https://go.dev/dl/go1.20.14.windows-amd64.msi)
3. Please also download SQLite Studio and install:
[SQLite Studio Download Link](https://sqlitestudio.pl/)
4. Once SQLite Studio is installed, open this (listener) directory in Windows Explorer, and open the /db folder. Copy and paste the template-kamar-directory-service.db file, then rename it to kamar-directory-service.db.
5. Open SQLite Studio, then in the top left click Add Database, and add kamar-directory-service.db. Connect to it by double-clicking on it in the sidebar on the left. Double click on the Results table - you should now see its structure in the main window. Click on the data tab - for now it will be empty, but this is where data from KAMAR will be populated. You will most likely need to click the Refresh button (blue icon) to see new data once an upload has been completed.
6. Create a folder in C:\\KamarListener\ called "tls". Generate key.pem and cert.pem files - [instructions can be found here](https://medium.com/@yakuphanbilgic3/create-self-signed-certificates-and-keys-with-openssl-4064f9165ea3). Move key.pem and cert.pem into the tls folder you created.
7. Make a copy of the .envrc.template file, rename it to .envrc, and populate it appropriately, using values that reflect those of your instance of KAMAR.
**IMPORTANT**
If you're testing or building the source code on your own computer, you need GCC installed and in your PATH. This is required in order for the Go program to connect to SQLite. If you are on Windows, the easiest way to do this is to install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/articles/2021-05/10.3.0-release).
**Also important**
When running `go run ./cmd/api` to test the project, you first need to ensure that CGO is enabled. To do this in PowerShell, run the command `$env:CGO_ENABLED=1`. This will ensure CGO is enabled for the lifespan of that PowerShell instance.

### To build a new .exe binary
1. Ensure that [TDM-GCC](https://jmeubank.github.io/tdm-gcc/articles/2021-05/10.3.0-release) is installed on your computer.
2. Open PowerShell, navigate to this directory, and use the following command:
`set CGO_ENABLED=1&& set goos=windows&& go build -o ./bin/kamarRefresh.exe ./cmd/api`
3. Copy this file (which will be in the /bin/ folder in this directory) across to (in this case) Mark's computer, to the C:\\KamarListener\ directory. Also copy the tls folder from this directory into the KamarListener directory. kamarRefresh.exe can then be double-clicked to run - a new terminal will open up, which will log any connections that the server receives.

### To run the program in dev mode
Open terminal, navigate to this directory, and use the following command:
`go run ./cmd/api`
The app runs on port 443 by default - consider including "-port 8084" or similar at the end of the command if there is an issue binding to port 443.
Use ctrl+c in the terminal to end the process.

### Optional flags for terminal commands
To see a list of available flags that can be used in either of the above terminal commands, use the following command in this directory:
`go run ./cmd/api -help`

### To test endpoints
If the service is running,  in the terminal navigate to the /test directory inside this folder (from the same computer that the application is running on). Use the following command:
`curl -k -X POST -d @./**filename**.json -H "Content-Type: application/json" --user "**username**:**password**" localhost:443/kamar-refresh`
You may need to use https://localhost... if HTTPS is turned on. The "-k" flag prevents the request from failing due to the TLS cert being self-signed (only necessary if HTTPS is turned on).

### Database
As only new results will be dumped each day, an SQLite database is used to hold all data, which PowerBI can connect to and consume.
<!-- 
### Testing - 10-4-24
- Had to use desktop IP address rather than localhost as address, and had to remember URL tag, eg. 192.168.1.84/kamar-refresh
- When trying Check and Enable, server displayed the error "failed at authCredentials", and also logged "received and processed check request". KAMAR displayed the error "ERROR: No service name returned"
- When trying Check and Enable with incorrect credentials (username:pa55word), server logged "failed at authCredentials" again, but this time didn't display "received and processed...". KAMAR displayed "ERROR: HTTP/1.1 403 Forbidden"
- Got past "failed at authCredentials" by adding a couple more fields to the SMSDirectoryData field of the response. Now, KAMAR showing this error: "ERROR: Invalid Server - invalid/missing support info URL. Please contact the supplier to update."
- KAMAR check now working - cause of areas was missing fields in the server response to KAMAR's check request. Final thing to add was a privacy statement that was more than just "none".

### Testing - 11-4-24
- TODO: Change port, as 443 might be too open by default, and is used for fmtp (required?) -->

### Expanding listening service
- data/ - create new ___.go to represent the new field. Use results.go as a template
- models.go - add new model
- refresh.go - create ___Field structs for each new field
- database.go - create new table statement