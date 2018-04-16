![FK](https://i.imgur.com/LozuBJ1.png)

# filekeep

[![GoDoc](https://godoc.org/github.com/vlad-s/filekeep?status.svg)](http://godoc.org/github.com/vlad-s/filekeep)
[![Go Report Card](https://goreportcard.com/badge/github.com/vlad-s/filekeep)](https://goreportcard.com/report/github.com/vlad-s/filekeep)
[![Twitter URL](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/0x766c6164)

> You own web file keeper

filekeep aims to be a simple, yet powerful, web file manager. It's easy to configure, and easier to use.
It has all assets (CSS files and HTML templates) bundled into the binary, for easy and instant deployment.

## Getting started

Assuming you have a working Go environment, installing and running it is as easy as doing:

```bash
go get https://github.com/vlad-s/filekeep && filekeep
```

This will get the latest version of `filekeep` and run it in the current directory.
For customizing the config, please read [Configuration](#configuration).

## Developing / Building

In case you want to modify the templates or CSS files without getting into the code, there's a little bash
script to help you rebuild the assets.

```bash
git clone https://github.com/vlad-s/filekeep
cd $GOPATH/src/github.com/vlad-s/filekeep
make
```

The `Makefile` provided has rules for running the binary (`make run` - also rebuilds the assets),
building the project (default, `make` or `make build`), (re)building the assets (`make assets`), and
writing the default config to disk (`make config` - builds & invokes `filekeep -dump-config`).

Building all assets can be done directly through executing `build_assets.sh` (or `make assets`), or for a single one,
provide the parameters directly into the shell as arguments. For example:

```bash
./build_assets.sh "templates:header.html:HTMLHeader"
```

### Deploying / Publishing

Usually you'll want to listen to a local port, and proxy the trafic through your web server of choice.
`filekeep` is not intended to be exposed directly to the internet, as there's no SSL/TLS configuration.
For example, on `nginx`, a simple proxy config for a local listener on port `8080` would look like this:

```
server {
    listen 80;
    listen [::]:80;

    server_name filekeep.domain.tld;

    return 301 https://filekeep.domain.tld$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/privkey.pem;
    
    server_name filekeep.domain.tld;
    
    location / {
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Host $http_host;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_pass http://localhost:8080;
    }
}
```

This will redirect all HTTP requests from your domain to HTTPS, which will proxy the traffic to the local listening port
of `filekeep`. For info on good SSL configuration for your web server of choice, please see [cipherli.st](https://cipherli.st/).

## Features

What can `filekeep` do?

* Extendable config, allowing fine tuning and many configuration possibilities:
    * Listening address and port,
    * Root directory,
    * Hiding files and directories by name, path, extension, or if starting with dot.
* Fast serving/routing, as result of [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter).
* Fast & instant deploy of the binary, as all assets are bundled:
    * Rebuilding the `.go` files of the assets is as easy as running `make assets`.
* Pretty logging, courtesy of [Sirupsen/logrus](https://github.com/Sirupsen/logrus):
    * The debugging flag in the config, if set to `true`, sets the debug level to **Debug**, otherwise defaults to **Info**.
* Breadcrumbs for easy navigation.
* Children files and directories count, file size.

## Configuration

Configuration of the project is managed by a `config.json` file next to the binary.
The default config (containing all options) can be dumped to disk using the `-dump-config` flag.
A local, customised config can be loaded at start up by using the `-load-config` flag.

#### -dump-config
Type: `bool`
Default: `false`

Will dump the default config to disk and exit. Logs any error encountered, if any.

Example:
```bash
filekeep -dump-config # dump the default config
ls config.json >/dev/null && echo "Config is present" # prints Config is present
```

#### -load-config
Type: `bool`
Default: `false`

Will load the config from disk, and continue to run. Logs any error encountered, if any.

Example:
```bash
filekeep -dump-config # dump the default config
$EDITOR config.json   # edit the config using your editor
filekeep -load-config # load the config
```

## Contributing

If you'd like to contribute to the development of `filekeep`, please fork the repository.
Pull requests are more than welcome.

For the contributing guide, please read the `CONTRIBUTING.md`
file inside this repo.

## Links

- Issue tracker: https://github.com/vlad-s/filekeep/issues
  - In case of sensitive bugs like security vulnerabilities, please contact vlad@vlads.me directly
    instead of using issue tracker. Thank you for improving the security and privacy of this project!


## Licensing

The code in this project is licensed under MIT license. Full license can be found in the `LICENSE` file
inside this repo.
