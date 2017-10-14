package sciencexml

// "crypto/tls"
// "hash"
// "mime/multipart"
// "net/url"
// "bytes"
// "crypto/rand"
// "crypto/sha512"
// "encoding/base64"
// "github.com/ziutek/mymysql/mysql"
// _ "github.com/ziutek/mymysql/native" // Native engine
// "html"
// "io"
// "strconv"
// "strings"
// "time"
import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
)

func fetchXML(webAddress, destinationFile string) {
}

func parseXML(w http.ResponseWriter, sourceFile string) {
	var content string
	filename := sourceFile
	fmt.Println("filename", filename)
	fmt.Fprint(w, "filename:", filename)
	xmlFile, err := os.Open(filename)
	fmt.Println("xmlFile", xmlFile)
	fmt.Fprint(w, "xmlFile:", xmlFile)
	fmt.Println("err", err)
	fmt.Fprint(w, "err:", err)
	if err != nil {
		fmt.Println(err)
		fmt.Fprint(w, err)
		return
	}
	decoder := xml.NewDecoder(xmlFile)
	fmt.Println("decoder", decoder)
	fmt.Fprint(w, "decoder", decoder)
	stack := make([]string, 24)
	sp := 0
	for {
		// Read tokens from the XML document in a stream.
		token, _ := decoder.Token()
		fmt.Println("token", token)
		fmt.Fprint(w, "token", token)
		if token == nil {
			break
		}
		// Inspect the type of the token just read.
		switch se := token.(type) {
		case xml.StartElement:
			fmt.Println("StartElement")
			fmt.Fprint(w, "StartElement")
			fmt.Println("se", se)
			fmt.Fprint(w, "se", se)
			name := se.Name.Local
			stack[sp] = name
			sp++
			fmt.Println("stack", stack)
			fmt.Fprint(w, "stack", stack)
			fmt.Println("sp", sp)
			fmt.Fprint(w, "sp", sp)
			content = ""
		case xml.EndElement:
			fmt.Println("EndElement")
			fmt.Fprint(w, "EndElement")
			fmt.Println("se", se)
			fmt.Fprint(w, "se", se)
			sp--
			name := se.Name.Local
			if stack[sp] != name {
				fmt.Println("Error: tag mismatch.")
				fmt.Fprint(w, "Error: tag mismatch")
				return
			}
			fmt.Println("stack", stack)
			fmt.Fprint(w, "stack", stack)
			fmt.Println("sp", sp)
			fmt.Fprint(w, "sp", sp)
			fmt.Println("content!!", content)
			fmt.Fprint(w, "content!!", content)
		case xml.CharData:
			fmt.Println("CharData")
			fmt.Fprint(w, "CharData")
			fmt.Println("se", se)
			fmt.Fprint(w, "se", se)
			chardata := string(se)
			fmt.Println(chardata)
			fmt.Fprint(w, chardata)
			content = content + chardata
		case xml.Comment:
			fmt.Println("Comment")
			fmt.Fprint(w, "Comment")
		case xml.ProcInst:
			fmt.Println("ProcInst")
			fmt.Fprint(w, "ProcInst")
			fmt.Println("se", se)
			fmt.Fprint(w, "se", se)
		case xml.Directive:
			fmt.Println("Directive")
			fmt.Fprint(w, "Directive")
			fmt.Println("se", se)
			fmt.Fprint(w, "se", se)
		}
	}
}
