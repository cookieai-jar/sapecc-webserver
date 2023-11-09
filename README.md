To build locally

> cd sap-app

> mvn clean compile assembly:single

The build webserver will listen on 9443 as the secure port and 9090 as the http port.

To start the webserver locally:

> java -cp sapjco3.jar:sap-app-1.0-SNAPSHOT-jar-with-dependencies.jar com.veza.app.App
