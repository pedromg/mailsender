[![Build Status](https://travis-ci.org/pedromg/mailsender.svg?branch=master)](https://travis-ci.org/pedromg/mailsender)
mailsender
=========

Very simple Go app to send notification email. Usually invoked by scripts.
A set of standard emails can be built using the JSON formatted configuration file. 
So several notifications emails can be pre-defined and used accordingly.

Each email will have a JSON file and can be stored anywhere. If a param with the 
`path/filename` is not passed, the default app path/files folder will be used to 
search the named JSON file.

Most of the config variables may be set via app param. 
_Attention_: calling params will overide JSON file configs.

Default config file is ```./mailsender.json```

The ```message``` can be passed into the ```body```template is this has a _%s_ or _%d_.

### Configuration file format

Format file: JSON structured:
```json
	{
		"app_name": "Alpha",
		"app_name": "monit",
		"secure": true,
		"smtp_host": "mailtrap.io",
		"smtp_port": 2525,
		"smtp_email": "sender@example.com",
		"smtp_username": "smtpusername123",
		"smtp_password": "123456",
		"email_address": "me@example.com", 
		"subject": "lift-off notification", 
		"body": "%s ready for lift-off",
		"message": "Eagle 3",
		"log": true, 
		"log_file": "mailsender_1.log",
	}
```

### Usage

```bash
$ mailsender -p /path/to/file -f filename.json
```


### Configuration fields

-   __configs__: (string) path to the JSON configs file
-	__server_name__: (string) server name that triggers the notification
-	__app_name__: (string) app name that triggers the notification
-	__secure__: (bool) http vs https.
-	__smtp_host__: (string) the hostname of the email provider
-	__smtp_port__: (int) the port of the smtp host
-	__smtp_email__: (string) the email of the sender (from header)
-	__smtp_username__: (string) the username for the smtp auth
-	__smtp_password__: (string) the password for the smtp auth
-	__email_address__: (string) email to receive the alerts.
-   __subject__: (string) a default subject for the email
-   __body__: (string) a default body message, accepting [%s|%d|...] params
-	__message__: (string) message passed into the app (composing the body)
-	__log__: (bool) log ?
-	__log_file__: (string) # file to append the log.

### Cross Compile

If you are building on OSX for Linux usage, make sure your Go e prepared to generate binaries for other architectures. To enable it for Linux:

```
$ cd  $GOROOT/src
$ GOOS=linux GOARCH=386 ./make.bash
```
Then to generate a linux specific binary:
```
$ GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o mailsender.linux mailsender.go
```

