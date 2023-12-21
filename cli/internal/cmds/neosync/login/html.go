package login_cmd

import (
	"fmt"
	"html/template"
	"io"
)

const (
	header = `
<!DOCTYPE html>
<head>
    <title>{{ .Title }}</title>
    <link rel="icon" type="image/png" href="https://assets.nucleuscloud.com/neosync/newbrand/favicon_transparent.ico" />
    <style>
        body {
            background-color: #101010;
        }

        .header {
            background-color: #FBFDF4;
        }

        .logo {
            width: 40%;
            padding-top: 20px;
            display: block;
            brightness:0.75;
            margin-left: auto;
            margin-right: auto
        }

        h1 {
            font-family:'Inter', sans-serif;
            font-size: 36px;
            letter-spacing: 1px;
            word-spacing: 2px;
            color: #EBECFA;
            font-weight: normal;
            text-transform: capitalize;
            text-align: center;
                  padding-top: 20px;
        }

        p {
            font-family:'Inter', sans-serif;
            font-size: 16px;
            letter-spacing: 1px;
            word-spacing: 2px;
                     color: #EBECFA;
            font-weight: normal;
            text-transform: normal;
            text-align: center;
        }

        .neosyncLogo {
            height: '40px';
            width: 40px;
        }

        .nav {
            border: 2px solid rgba(117, 117, 117, 1);
            border-radius: 10px;
            padding: 10px;
        }

        #content {
            padding:10px
        }

        #footer {
            font-size: 0.8em;
        }

        .error-text {
            font-weight: bold;
        }
    </style>
</head>

<body>
    <div id="content">
	`

	footer = `
	</div>
  <div id="footer"></div>
</body>
</html>
	`

	loginPageSuccess = `
  <div class='nav'><a href="https://neosync.dev"><img class='neosyncLogo' src="https://assets.nucleuscloud.com/neosync/newbrand/logo_light_mode.svg"></a></div>
  <div class='successText'>
      <h1>Login Success!</h1>
      <p>You've successfully logged in to Neosync CLI.</p>
      <p>You may now close this window and return to your terminal.</p>
  </div>
  <div>
      <img class='logo' src="https://assets.nucleuscloud.com/neosync/app/cliImage.png">
  </div>
	`

	loginPageError = `
    <div><a href="https://neosync.dev"><img class='neosyncLogo' src="https://assets.nucleuscloud.com/neosync/newbrand/logo_light_mode.svg"></a></div>
    <div class='successText'>
        <h1>There was a problem logging you in!</h1>
        <p class="error-text">Error Code: {{ .ErrorCode }}</p>
        <p class="error-text">Error Description: {{ .ErrorDescription }}</p>
    </div>
    <div>
        <img class='logo' src="https://assets.nucleuscloud.com/neosync/app/angryDarth.jpg">
    </div>
	`
)

// wraps page with header and footer
func wrapPage(contents string) string {
	return fmt.Sprintf(
		`
{{ template "header" . }}
%s
{{ template "footer" . }}
`, contents,
	)
}

type loginPageData struct {
	Title string
}

func renderLoginSuccessPage(wr io.Writer, data loginPageData) error {
	pageTmpl, err := getHtmlPage()
	if err != nil {
		return err
	}
	pageTmpl, err = pageTmpl.New("login").Parse(wrapPage(loginPageSuccess))
	if err != nil {
		return err
	}
	return pageTmpl.ExecuteTemplate(wr, "login", data)
}

type loginPageErrorData struct {
	Title string

	ErrorCode        string
	ErrorDescription string
}

func renderLoginErrorPage(wr io.Writer, data loginPageErrorData) error {
	pageTmpl, err := getHtmlPage()
	if err != nil {
		return err
	}
	pageTmpl, err = pageTmpl.New("login").Parse(wrapPage(loginPageError))
	if err != nil {
		return err
	}
	return pageTmpl.ExecuteTemplate(wr, "login", data)
}

// returns a template with the header and footer templates added in
func getHtmlPage() (*template.Template, error) {
	templ, err := template.New("header").Parse(header)
	if err != nil {
		return nil, err
	}
	templ, err = templ.New("footer").Parse(footer)
	if err != nil {
		return nil, err
	}
	return templ, nil
}
