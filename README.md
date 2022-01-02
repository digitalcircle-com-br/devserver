# devserver

Local HTTPS Server with tools to ease development

Made to all of us, speding hard time to develop locally due to Cookies, CORS, HTTPS reinforcement - well, this tool is
for you.

Devserver is a http gateway to simplify usage on you machine.

Devserver allows creating of virtual hosts locally, also managing the certificates

> Video in portuguese explaining it a little: [LINK](https://drive.google.com/file/d/1mxtEHXyhn09WPiYBrEyavhpsABXgR2WW/view?usp=sharing)
## Releases

Please check here: [Releases](https://github.com/digitalcircle-com-br/devserver/releases)

## How to use it

Install Devserver

> go install github.com/digitalcircle-com-br/devserver/cmd/devserver@latest

Once done, it will generate a .devserver folder in your home dir, also initiating a CAROOT in there.

> Important: add the file <HOME>/.devserver/caroot/ca.cer to your pool of trusted certificates.

Now its time to configure it:

```yaml
addr: ":8443"
log: web
routes:
  app.dev.local/static/: static:~/app/static
  app.dev.local/raw/: raw:~DS/devserver/raw
  api.dev.local: http://localhost:8082
```

> Special path subs: ~ will be replaced as userhome; ~DS will be replaced with devserver home

Based on the config above, we can find the following:

1 - addr: is the address/port the gateway will listen.
2 - log: how to track log. In case value is `-`, will be sent to stdout. in case value is `web`, we will use weg log interface. Everything else will be considered a file name, and log will be sent there. For the log web interface, either open you browser on `https://localhost<addr>/__log/index.html` or use the provided menu in tray.
3 - routes: define different virtual hosts / paths to address your requests

A route is always a hostname and path, being the minimal path "/". Its important to define a hostname "*" to be used as
default host in case request does not match any previously set route.

The value of a route is a string with procotol://direction.

Protocol may be:

- static: in case you want to serve static pages from the filesystem
- raw: similar to static, but the files are raw http responses - suppose a mock api. If you add _GET for example to the
  name of the file it will be used only to reply to GET request.s
- serverless: allows usage of serverless programs (like cgi in the old times), please refer to serverless section ahead.
- http/https are standard protocols, and will act as reverse proxy for these endpoints. 3 - log: defines if will log in
  file w rotate. File name is given here. If file name is  "-", then stdout will be used.


# Serverless

Serverless are programs that will read a request from stdin, process and send a raw http respose through stdout.

We propose reading: [Nanoserverless](https://github.com/digitalcircle-com-br/nanoserverless) for deeper understaning.

The core idea here is: serverless://`<dir>` will make dir the root of a serverless fs tree. Requests will have src path prefix stripped, and url path will map to a file inside dir, named `<strippedpath>.yaml` or `<strippedpath>_METHOD.yaml`.
In case the file with _METHOD is found, will have higher priority over `<strippedpath>.yaml`.

The file content is like:

```yaml
cmd: go
params: [run, ./serverless/a_get.go]
```
In this way, devserver will find what to execute upon request. the cmd with the params will be started, request will be sent to ti through stdin. Response will be sent back through stdout.

A sample serverless program in go, using nanoserverless is like:

```go
package main

import (
	"net/http"

	"github.com/digitalcircle-com-br/nanoserverless"
)

func main() {
	nanoserverless.ServeSIO(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Hello World"))
	})
}
```
> Keep in mind paths are relative to ~/.devserver when setting up your serverless trees
# Using virtual hosts

To keep adherend, we strongly suggest the usage of virtual hosts, by editing /etc/host or c:
\windows\system32\drivers\etc\hosts

# Install

Please consider getting it from the releases page. In case your arch is not there, clone the repo and use the makefile (read it 1st, ok?)

In case you want to contribute by making the bins to your platform, please join the band

After that, just launch it, on menu, ask to open dir and edit your config.

# Complementary Setup
 - [Editing Hosts File](https://linuxize.com/post/how-to-edit-your-hosts-file/)
 - [Add Certificate Windows](https://support.securly.com/hc/en-us/articles/360026808753-How-do-I-manually-install-the-Securly-SSL-certificate-on-Windows)
 - [Add Certificate Mac](https://support.securly.com/hc/en-us/articles/206058318-How-to-install-the-Securly-SSL-certificate-on-Mac-OSX-)
 - [Add Certificate Linux](https://askubuntu.com/questions/645818/how-to-install-certificates-for-command-line)

Have fun!