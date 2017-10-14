package youtube

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"static"
	"strconv"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
)

const missingClientSecretsMessage = `
Please configure OAuth 2.0
`

// var (
//         clientSecretsFile = flag.String("secrets", "cs.json", "Client Secrets configuration")
//         cacheFile         = flag.String("cache", "request.token", "Token cache file")
// )

// Our config info
func getClientSecretsFile() string {
	return "/home/ec2-user/wayneserver/cs.json"
}

func getCacheFile() string {
	return "/home/ec2-user/wayneserver/request.token"
}

// ClientConfig is a data structure definition for the client_secrets.json file.
// The code unmarshals the JSON configuration file into this structure.
type ClientConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURIs []string `json:"redirect_uris"`
	AuthURI      string   `json:"auth_uri"`
	TokenURI     string   `json:"token_uri"`
}

// Config is a root-level configuration object.
type Config struct {
	Installed ClientConfig `json:"installed"`
	Web       ClientConfig `json:"web"`
}

// openURL opens a browser window to the specified location.
// This code originally appeared at:
//   http://stackoverflow.com/questions/10377243/how-can-i-launch-a-process-that-is-not-a-file-in-go
func openURL(url string) error {
	fmt.Println("openURL")
	fmt.Println("url", url)
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:4001/").Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("Cannot open URL %s on this platform", url)
	}
	return err
}

// readConfig reads the configuration from clientSecretsFile.
// It returns an oauth configuration object for use with the Google API client.
func readConfig(scope string) (*oauth2.Config, error) {
	// Read the secrets file
	clientSecretsFile := getClientSecretsFile()
	data, err := ioutil.ReadFile(clientSecretsFile)
	if err != nil {
		pwd, _ := os.Getwd()
		fullPath := filepath.Join(pwd, clientSecretsFile)
		return nil, fmt.Errorf(missingClientSecretsMessage, fullPath)
	}

	cfg := new(Config)
	fmt.Println("data", data)
	fmt.Println("cfg", cfg)
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	var redirectUri string
	if len(cfg.Web.RedirectURIs) > 0 {
		redirectUri = cfg.Web.RedirectURIs[0]
		fmt.Println("redirectUri", redirectUri)
	} else if len(cfg.Installed.RedirectURIs) > 0 {
		redirectUri = cfg.Installed.RedirectURIs[0]
		fmt.Println("redirectUri", redirectUri)
	} else {
		return nil, errors.New("Must specify a redirect URI in config file or when creating OAuth client")
	}

	return &oauth2.Config{
		ClientID:     cfg.Installed.ClientID,
		ClientSecret: cfg.Installed.ClientSecret,
		Scopes:       []string{scope},
		Endpoint:     oauth2.Endpoint{cfg.Installed.AuthURI, cfg.Installed.TokenURI},
		RedirectURL:  redirectUri,
	}, nil
}

// startWebServer starts a web server that listens on http://localhost:8080.
// The webserver waits for an oauth code in the three-legged auth flow.
func startWebServer() (codeCh chan string, err error) {
	fmt.Println("startWebServer")
	fmt.Println("codeCh", codeCh)
	fmt.Println("err", err)
	listener, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		return nil, err
	}
	codeCh = make(chan string)
	go http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.FormValue("code")
		codeCh <- code // send code to OAuth flow
		listener.Close()
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Received code: %v\r\nYou can now safely close this browser window.", code)
	}))

	return codeCh, nil
}

// buildOAuthHTTPClient takes the user through the three-legged OAuth flow.
// It opens a browser in the native OS or outputs a URL, then blocks until
// the redirect completes to the /oauth2callback URI.
// It returns an instance of an HTTP client that can be passed to the
// constructor of the API client.
func buildOAuthHTTPClient(scope string) (*http.Client, error) {
	config, err := readConfig(scope)
	if err != nil {
		msg := fmt.Sprintf("Cannot read configuration file: %v", err)
		return nil, errors.New(msg)
	}
	fmt.Println("config", config)

	var ctx context.Context

	// Try to read the token from the cache file.
	// If an error occurs, do the three-legged OAuth flow because
	// the token is invalid or doesn't exist.
	var token *oauth2.Token

	cacheFile := getCacheFile()
	data, err := ioutil.ReadFile(cacheFile)
	if err == nil {
		err = json.Unmarshal(data, &token)
	}
	if (err != nil) || !token.Valid() {
		// Start web server.
		// This is how this program receives the authorization code
		// when the browser redirects.
		codeCh, err := startWebServer()
		if err != nil {
			return nil, err
		}
		fmt.Println(codeCh)

		// Open url in browser
		url := config.AuthCodeURL("")
		err = openURL(url)
		if err != nil {
			fmt.Println("Visit the URL below to get a code.",
				" This program will pause until the site is visted.")
		} else {
			fmt.Println("Your browser has been opened to an authorization URL.",
				" This program will resume once authorization has been provided.\n")

		}
		// Accept code on command line.
		fmt.Println(url)
		fmt.Print("Enter code: ")
		scanner := bufio.NewScanner(os.Stdin)
		code := ""
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
			code = line
			break
		}
		fmt.Println("code", code)

		// This code caches the authorization code on the local
		// filesystem, if necessary, as long as the TokenCache
		// attribute in the config is set.
		token, err = config.Exchange(ctx, code)
		if err != nil {
			return nil, err
		}
		fmt.Println("token", token)
		data, err := json.Marshal(token)
		ioutil.WriteFile(cacheFile, data, 0644)
	}

	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(token)), nil
}

func handleError(err error, message string) {
	if message == "" {
		message = "Error making API call"
	}
	if err != nil {
		log.Fatalf(message+": %v", err.Error())
	}
}

func addPropertyToResource(ref map[string]interface{}, keys []string, value string, count int) map[string]interface{} {
	for k := count; k < (len(keys) - 1); k++ {
		switch val := ref[keys[k]].(type) {
		case map[string]interface{}:
			ref[keys[k]] = addPropertyToResource(val, keys, value, (k + 1))
		case nil:
			next := make(map[string]interface{})
			ref[keys[k]] = addPropertyToResource(next, keys, value, (k + 1))
		}
	}
	// Only include properties that have values.
	if count == len(keys)-1 && value != "" {
		valueKey := keys[len(keys)-1]
		if valueKey[len(valueKey)-2:] == "[]" {
			ref[valueKey[0:len(valueKey)-2]] = strings.Split(value, ",")
		} else if len(valueKey) > 4 && valueKey[len(valueKey)-4:] == "|int" {
			ref[valueKey[0:len(valueKey)-4]], _ = strconv.Atoi(value)
		} else if value == "true" {
			ref[valueKey] = true
		} else if value == "false" {
			ref[valueKey] = false
		} else {
			ref[valueKey] = value
		}
	}
	return ref
}

func createResource(properties map[string]string) string {
	resource := make(map[string]interface{})
	for key, value := range properties {
		keys := strings.Split(key, ".")
		ref := addPropertyToResource(resource, keys, value, 0)
		resource = ref
	}
	propJson, err := json.Marshal(resource)
	if err != nil {
		log.Fatal("cannot encode to JSON ", err)
	}
	return string(propJson)
}

func printSearchListResults(response *youtube.SearchListResponse) {
	for _, item := range response.Items {
		fmt.Println(item.Id.VideoId, ": ", item.Snippet.Title)
		fmt.Println("item", item)
	}
}

func searchList(service *youtube.Service) {
	call := service.Search.List("snippet")
	call = call.MaxResults(2)
	// call = call.Q(q)
	call = call.Type("video")
	call = call.Order("date")
	response, err := call.Do()
	handleError(err, "")
	printSearchListResults(response)
}

func doSomething() {
	// flag.Parse()

	client, err := buildOAuthHTTPClient(youtube.YoutubeForceSslScope)
	handleError(err, "Error building OAuth client")

	service, err := youtube.New(client)
	handleError(err, "Error creating YouTube client")

	searchList(service)
}

func getDoctype() string {
	return `<!DOCTYPE html>
<html>
<head>
<meta charset=utf-8 />
`
}

func showListPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	doSomething()
}

func showOAuthRedirectPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>YouTube OAuth2 Redirect Page</title>
<link rel="stylesheet" type="text/css" href="/style.css">
</head>
<body>
<section>
<h1>YouTube OAuth2 Redirect Page</h1>
This page apparently has to exist, but I don't know why.
</section>
</body>
</html>`)
}

func Handler(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	fmt.Println("We are in the youtube handler, and op is", op)
	switch {
	case op == "list":
		showListPage(w, r, op, userid, userName)
	case op == "oauth2":
		showOAuthRedirectPage(w, r, op, userid, userName)
	default:
		// fmt.Fprintln(w, "Could not find page:", op)
		filename := "/home/ec2-user/wayneserver/staticappcontent/youtube/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}
