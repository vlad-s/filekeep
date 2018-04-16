# Contributing

## Go files

* All files must be formatted using `gofmt` and `goimports`, nothing else in particular.

## Assets

Assets refer to any non-go file inside the project, that will be bundled
into the binary for easy and instant deployment.

* After editing any of the assets files, run `make assets`.

The `build_assets.sh` file will take care of embedding code in Go constants.

### Adding a new asset file

`build_assets.sh` has a little function called `build` which will do the hard
work of creating the Go file which will embed the asset file's code into the binary.

`build` has three parameters: `filePath package constName`. All arguments are stored
into an array of the form `directory:filename.ext:constName`.

For example, let's add a custom CSS file:
```bash
cd $GOPATH/src/github.com/vlad-s/filekeep # cd into the project's directory
touch assets/css/my-custom.css   # create the CSS file
$EDITOR assets/css/my-custom.css # edit the CSS file
$EDITOR build_assets.sh          # append the file to the assets array
cat build_assets.sh              # output reduced for readability
[...]
FILES=(
    css:hack.css:HackCSS
    css:custom.css:CustomCSS
    css:my-custom.css:MyCustomCSS # <- newly introduced file

    templates:404.html:HTML404
    templates:dir.html:HTMLDirList
    templates:header.html:HTMLHeader
    templates:footer.html:HTMLFooter
)
[...]
make assets # alternatively, `./build_assets.sh`
```

Afterwards, you can easily use `assets.MyCustomCSS` in your Go code to embed the content.
