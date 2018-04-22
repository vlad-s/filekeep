package templates

/*
DO NOT EDIT
Autogenerated file by `build_assets.sh` at Sun Apr 22 18:53:09 EEST 2018.
*/

// HTMLPassForm - bundled asset, name should be self explaining.
const HTMLPassForm = `
{{template "header"}}

<div class="container">
    <div class="grid">
        <div class="cell -12of12">
            <div class="card">
                <header class="card-header">
                    enter password for file <strong>{{.Name}}</strong>
                </header>
                <div class="card-content">
                    <div class="inner">
                        <div class="grid -center">
                            <div class="cell -4of12">
                                <form class="form" method="post">
                                    <fieldset class="form-group form-warning">
                                        <label for="password">pass:</label>
                                        <input id="password" name="password" type="password" class="form-control">
                                    </fieldset>
                                </form>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>

{{template "footer"}}
`
