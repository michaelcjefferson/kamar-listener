# KAMAR Listener
A secure, local-network, browser-based application that receives your school's data from KAMAR's Directory Services and stores them in a structured SQLite database.
This data can then be consumed by PowerBI or any other data analytics system used by your school's data team.

## To get started:
1. Download the [most recent release](https://github.com/michaelcjefferson/kamar-listener/releases).
2. Run listenerService.exe on a machine on the same local network as your instance of KAMAR. Click "More Info" --> "Run Anyway" --> "Yes" when prompted.
3. Open a browser and navigate to [https://localhost:8085](https://localhost:8085).
4. Conigure both a user account for KAMAR Listener and authentication details for Directory Services when prompted.
5. Open KAMAR, go to Directory Services in the Setup menu, and fill in the required details.
6. Press Check & Enable.  

[Detailed set-up instructions below](#setting-things-up).

## Privacy
This application is configured to run on a local network - all data remains local, and the application is not publicly accessible unless your school's IT team forwards the application port to the outside world.

In case students are using the same internal network as your device, robust user authentication has been set up for KAMAR Listener, and the database itself is not viewable through the application - the weakest entrypoint is the Basic Auth used by Directory Services itself, so ensure you use a complex and secure password for this.

Directory Services can include sensitive information such as NSNs as part of its service, and these are written to a .db file. Please ensure that KAMAR Listener is run on a secure device.

## Setting things up
These steps assume you are running listenerService.exe on a Windows machine, which is on the same local network as your instance of KAMAR.
1. Download the [most recent release](https://github.com/michaelcjefferson/kamar-listener/releases) of KAMAR Listener and **run it**. Windows will display a warning:
<img src="/ui/assets/unrecognised-app-warning.png" alt="Windows Unrecognised App warning" width="450">  
To enable the application, click on More Info, and you will see the following (double-check that it says "Application: listenerService.exe"):
<img src="/ui/assets/unrecognised-app-expanded.png" alt="Windows Unrecognised App warning expanded" width="450">  
Click "Run Anyway" to allow the application to run on your device.


2. KAMAR Directory Services requires an HTTPS connection for security. To enable this, KAMAR Listener automatically downloads [mkcert](https://github.com/FiloSottile/mkcert) to generate and trust a local development certificate for use in this application. You will need to allow this when you first run KAMAR Listener:
<img src="/ui/assets/ca-warning.png" alt="Certificate Authority warning" width="450">  
Instead of HERBERT-THE-AVE\LENOVO@Herbert-the-Avenger, you will see your device name, which will act as the SSL certificate authority. Click "Yes" to approve - this is required in order for Directory Services to connect.

3. Run the listenerService.exe file - a terminal will open representing the running server. DO NOT CLOSE THIS - it will stop the listener service.

4. Open a browser and navigate to https://localhost:8085 - this is the user interface for this application. You will be prompted to create an admin account, allowing you to view logs, create new users, configure settings etc. for the listener service.
It will then prompt you to set up a username and password for Directory Services - these are the credentials you will also need to provide on the KAMAR Directory Services set-up page.

5. Two SQLite databases (.db files) are automatically created and used by this service:  
    - app.db, which stores information such as logs, user credentials etc. for KAMAR Listener
    - listener.db, which stores all data received from Directory Services

These databases perpetuate - if you restart your computer or end KAMAR Listener, this data will not be lost unless you delete these two database files. To view the data, click on "Open Database Folder" on the KAMAR Listener dashboard.

6. Open KAMAR, and go to Setup --> Server --> Directory Services. Fill out the details there (refer to [this page](https://directoryservices.kamar.nz/?listening-service) for help):  
    - Username and password need to be the same as the ones you set up (go to config page of KAMAR Listener to change)
    - Port needs to be 8085
    - **URL needs to be accurate**: "your IP address/kamar-listener", eg. `178.168.50.104/kamar-listener`. Your IP address should be visible on the dashboard - if this is not accurate, open a terminal and run `ipconfig` to find it.  

Click Check and Run, then tick required fields (eg. Results, setting it up with the timeframe required), then click Update to start the preliminary upload. You should see logs appearing in the terminal window running listenerService.exe, as well as rows of results starting to appear in SQLite Studio (need to click the blue refresh button on the Data tab).

7. If you want to end the listener service, go to the terminal that opened when you started listenerService.exe, and either close it or press Ctrl+C to stop the process. 

8. To restart the service, run listenerService.exe again.

In order to ensure security, this application does not connect to the world outside of your local network - if you want a remotely accessible service (not advised), this can be easily achieved by running it on a cloud instance (eg. [Google Cloud](https://console.cloud.google.com/)).

## Starting over/updating
To run an updated version of KAMAR Listener but keep your data intact, just delete the old version of listenerService.exe and download and run the new one - it will integrate with the databases that were previously created.


To start over with fresh, empty databases, open the database folder (via "Open Database Folder" on the dashboard), stop listenerService.exe if it is running, and delete **both** app.db **and** listener.db. Then, run listenerService.exe again.

## Development
### Setting up your dev environment
1. Clone a copy of this repository to your computer (or download it).
2. Download Go v1.23.6 (or a later version) from the following link and install on your computer:
[Go Download Link](https://go.dev/dl/go1.23.6.windows-amd64.msi). Open a terminal and run `go version` - if Go was installed successfully, you should see something similar to "go version go1.23.6 ...."
3. Install [Taskfile](https://taskfile.dev/) by running `go install github.com/go-task/task/v3/cmd/task@latest` - this tool simplifies the run and build processes. Run `task --version` in the terminal to ensure installation was successful.
4. Navigate to the /kamar-listener folder in your terminal, then run `task -a` to see a list of available commands (or check the Taskfile.yaml file in a code/text editor).
5. To run the application in development mode, use `task windows/run/api` (or `task run/api` if developing on Linux).

### To test endpoints
With the service running in one terminal, open another terminal. Navigate to /kamar-listener/test. Run `go run . -file *./path_to_file*`, replacing path_to_file with the path to the .json you want to send to KAMAR Listener.

### To run other tests
Navigate to /kamar-listener/cmd/api, and run `go test -v`.

### Expanding listening service
- database.go - create new table statement
- data/ - create new ___.go to represent the new field. Use results.go as a template
- models.go - add new model
- refresh.go - create ___Field structs for each new field, and add to the switch-case statement