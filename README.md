# devserver

Local HTTPS Server with tools to ease development

Made to all of us, speding hard time to develop locally due to Cookies, Cors, HTTPS reinforcement - well, this tool is
for you.

Devserver is a http gateway to simplify usage on you machine.

Devserver allows creating of virtual hosts locally, also managing the certificates

## How to use it

Install Devserver

> go install github.com/digitalcircle-com-br/devserver/cmd/devserver@latest

Once done, it will generate a .devserver folder in your home dir, also initiating a CAROOT in there.

> Important: add the file <HOME>/.devserver/caroot/ca.cer to your pool of trusted certificates.

Now its time to configure it:

```yaml
---
addr: ":8443"
log: devserver.log
routes:
  _app.dc.local: https://www.slashdot.org
  app.dc.local/s/: https://www.slashdot.org
  app.dc.local/g/: https://www.google.com
  app.dc.local/static/: static:///<somedir>/devserver
  app.dc.local/raw/: raw:///<somedir>/devserver/raw
  api.dc.local: http://localhost:8082
```

Based on the config above, we can find the following:

1 - addr: is the address/port the gateway will listen.

2 - routes: define different virtual hosts / paths to address your requests

A route is always a hostname and path, being the minimal path "/". Its important to define a hostname "*" to be used as
default host in case request does not match any previously set route.

The value of a route is a string with procotol://direction.

Protocol may be:

- static: in case you want to serve static pages from the filesystem
- raw: similar to static, but the files are raw http responses - suppose a mock api. If you add _GET for example to the
  name of the file it will be used only to reply to GET request.s
- http/https are standard protocols, and will act as reverse proxy for these endpoints. 3 - log: defines if will log in
  file w rotate. File name is given here. If file name is  "-", then stdout will be used.

# Using virtual hosts

To keep adherend, we strongly suggest the usage of virtual hosts, by editing /etc/host or c:
\windows\system32\drivers\etc\hosts

# Install
### Unix (Mac/Linux,etc)
  go install github.com/digitalcircle-com-br/devserver@latest
### Windows
  go install -ldflags "-H=windowsgui"  github.com/digitalcircle-com-br/devserver@latest
  
After that, just launch it, on menu, ask to open dir and edit your config.

Have fun!