package main

import (
	"encoding/json"
	"github.com/TylerBrock/colorjson"
	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Creator struct {
	Name    string  `json:"name"`
	Version string  `json:"version"`
	Comment *string `json:"comment"`
}

type Browser struct {
	Name    string  `json:"name"`
	Version string  `json:"version"`
	Comment *string `json:"comment"`
}

type PageTiming struct {
	ContentLoad *int    `json:"onContentLoad"`
	Load        *int    `json:"onLoad"`
	Comment     *string `json:"comment"`
}

type Page struct {
	StartedDateTime string       `json:"startedDateTime"`
	Id              string       `json:"id"`
	Title           string       `json:"title"`
	PageTimings     []PageTiming `json:"pageTimings"`
	Comment         *string      `json:"comment"`
}

type Cookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Path     *string `json:"path"`
	Domain   *string `json:"domain"`
	Expires  *string `json:"expires"`
	HttpOnly *bool   `json:"httpOnly"`
	Secure   *bool   `json:"secure"`
	Comment  *string `json:"comment"`
}

type Header struct {
	Name    string  `json:"name"`
	Value   string  `json:"value"`
	Comment *string `json:"comment"`
}

type QueryParameter struct {
	Name    string  `json:"name"`
	Value   string  `json:"value"`
	Comment *string `json:"comment"`
}

type PostParameters struct {
	Name        string  `json:"name"`
	Value       *string `json:"value"`
	FileName    *string `json:"fileName"`
	ContentType *string `json:"contentType"`
	Comment     *string `json:"comment"`
}

type PostData struct {
	MimeType string           `json:"mimeType"`
	Params   []PostParameters `json:"params"`
	Text     string           `json:"text"`
	Comment  *string          `json:"comment"`
}

type Content struct {
	Size        int     `json:"size"`
	Compression *int    `json:"compression"`
	MimeType    string  `json:"mimeType"`
	Text        *string `json:"text"`
	Encoding    *string `json:"encoding"`
	Comment     *string `json:"comment"`
}

type Response struct {
	Status      int      `json:"status"`
	StatusText  string   `json:"statusText"`
	HttpVersion string   `json:"httpVersion"`
	Cookies     []Cookie `json:"cookies"`
	Headers     []Header `json:"headers"`
	Content     *Content `json:"content"`
	RedirectUrl *string  `json:"redirectURL"`
	HeadersSize int      `json:"headersSize"`
	BodySize    int      `json:"bodySize"`
	Comment     *string  `json:"comment"`
}

type Request struct {
	Method      string           `json:"method"`
	Url         string           `json:"url"`
	HttpVersion string           `json:"httpVersion"`
	Cookies     []Cookie         `json:"cookies"`
	Headers     []Header         `json:"headers"`
	QueryString []QueryParameter `json:"queryString"`
	PostData    *PostData        `json:"postData"`
	HeaderSize  int              `json:"headerSize"`
	BodySize    int              `json:"bodySize"`
	Comment     *string          `json:"comment"`
}

type CacheState struct {
	Expires    *string `json:"expires"`
	LastAccess string  `json:"lastAccess"`
	ETag       string  `json:"eTag"`
	HitCount   int     `json:"hitCount"`
	Comment    *string `json:"comment"`
}

type Cache struct {
	BeforeRequest *CacheState `json:"beforeRequest"`
	AfterRequest  *CacheState `json:"afterRequest"`
	Comment       *string     `json:"comment"`
}

type EntryTimings struct {
	Blocked *int    `json:"blocked"`
	Dns     *int    `json:"dns"`
	Connect *int    `json:"connect"`
	Send    int     `json:"send"`
	Wait    int     `json:"wait"`
	Receive int     `json:"receive"`
	Ssl     *int    `json:"ssl"`
	Comment *string `json:"comment"`
}

type Entry struct {
	PageRef         *string      `json:"pageref"`
	StartedDateTime string       `json:"startedDateTime"`
	TimeMs          int          `json:"time"`
	Request         Request      `json:"request"`
	Response        Response     `json:"response"`
	Cache           Cache        `json:"cache"`
	Timings         EntryTimings `json:"timings"`
	ServerIP        *string      `json:"serverIPAddress"`
	Connection      *string      `json:"connection"`
	Comment         *string      `json:"comment"`
}

type Log struct {
	Version string   `json:"version"`
	Creator Creator  `json:"creator"`
	Browser *Browser `json:"browser"`
	Pages   *[]Page  `json:"pages"`
	Entries []Entry  `json:"entries"`
	Comment *string  `json:"comment"`
}

type HarFile struct {
	Log Log `json:"log"`
}

// Things to filter on
//   Request Domain (includes)
//   Request Path (includes)
//   Has Body
//   Response Status Code
// Things to print
//   Include headers
//   Include cookies
//   Include body
//   Include timings

// A       E F G   I J K L M N O   Q R S T     W X Y Z
// a       e   g h   j k l   n o   q r   t     w x y z

var CLI struct {
	RequestDomain         *string   `short:"D" name:"request-domain" help:"Find results where the domain equals this value"`
	RequestDomainIncludes *string   `short:"d" name:"request-domain-includes" help:"Find results where the domain contains this value"`
	RequestPath           *string   `short:"P" name:"request-path" help:"Find results where the request path equals this value"`
	RequestPathIncludes   *string   `short:"p" name:"request-path-includes" help:"Fina results where the request path includes this value"`
	RequestHasBody        *bool     `short:"b" name:"request-has-body" help:"Find results where the request has a body"`
	ResponseHasBody       *bool     `short:"B" name:"response-has-body" help:"Find results where the response has a body"`
	MethodIn              *[]string `short:"m" name:"method-in" help:"Find requests where the method is one of the provided values"`
	ResponseCode          *int      `short:"c" name:"response-code" help:"Find requests where the response code is equal to the value"`
	ResponseInformational *bool     `short:"i" name:"response-informational" help:"Find requests where the response was successful"`
	ResponseSuccessful    *bool     `short:"s" name:"response-success" help:"Find requests where the response was successful"`
	ResponseFailed        *bool     `short:"f" name:"response-fail" help:"Find requests where the responses was unsuccessful"`
	IncludeHeaders        *bool     `short:"H" name:"print-headers" help:"If specified, the request and response headers (excluding Cookie headers, use -C for that) will be included in the output"`
	IncludeCookies        *bool     `short:"C" name:"print-cookies" help:"If specified, the request and response cookies will be included in the output"`
	IncludeRequestBody    *bool     `short:"u" name:"print-request-body" help:"If specified, include the body of the request, including JSON highlighting"`
	IncludeResponseBody   *bool     `short:"U" name:"print-response-body" help:"If specified, include the body of the response, including JSON highlighting"`
	IncludeTimings        *bool     `short:"t" name:"print-timings" help:"If specified, include the request timings"`
	File                  string    `arg:"" help:"The HAR file to parse" type:"existingfile"`
}

func IsEntryValid(entry Entry) bool {
	requestUrl, err := url.Parse(entry.Request.Url)
	if err != nil {
		slog.Error("Failed to process url, ignoring request", "url", entry.Request.Url, "error", "err")
		return false
	}

	if CLI.RequestDomain != nil {
		if strings.ToLower(requestUrl.Host) != strings.ToLower(*CLI.RequestDomain) {
			return false
		}
	}
	if CLI.RequestDomainIncludes != nil {
		if !strings.Contains(strings.ToLower(requestUrl.Host), strings.ToLower(*CLI.RequestDomainIncludes)) {
			return false
		}
	}
	if CLI.RequestPath != nil {
		if strings.ToLower(requestUrl.Path) != strings.ToLower(*CLI.RequestPath) {
			return false
		}
	}
	if CLI.RequestPathIncludes != nil {
		if !strings.Contains(strings.ToLower(requestUrl.Path), strings.ToLower(*CLI.RequestPathIncludes)) {
			return false
		}
	}
	if CLI.RequestHasBody != nil {
		if *CLI.RequestHasBody {
			if entry.Request.BodySize == 0 {
				return false
			}
		} else {
			if entry.Request.BodySize > 0 {
				return false
			}
		}
	}
	if CLI.ResponseHasBody != nil {
		if *CLI.ResponseHasBody {
			if entry.Response.BodySize == 0 {
				return false
			}
		} else {
			if entry.Response.BodySize > 0 {
				return false
			}
		}
	}
	if CLI.MethodIn != nil {
		anyMatch := false
		for _, method := range *CLI.MethodIn {
			if strings.ToLower(entry.Request.Method) == strings.ToLower(method) {
				anyMatch = true
				break
			}
		}

		if !anyMatch {
			return false
		}
	}
	if CLI.ResponseCode != nil {
		if entry.Response.Status != *CLI.ResponseCode {
			return false
		}
	}
	if CLI.ResponseSuccessful != nil {
		if entry.Response.Status < 200 || entry.Response.Status > 399 {
			return false
		}
	}
	if CLI.ResponseInformational != nil {
		if entry.Response.Status < 100 || entry.Response.Status > 199 {
			return false
		}
	}
	if CLI.ResponseFailed != nil {
		if entry.Response.Status < 400 || entry.Response.Status > 599 {
			return false
		}
	}

	return true
}

func Filter[T interface{}](slice []T, predicate func(v T) bool) []T {
	result := make([]T, 0)
	for _, t := range slice {
		if predicate(t) {
			result = append(result, t)
		}
	}

	return result
}

func Tertiary[T interface{}](c bool, ok T, not T) T {
	if c {
		return ok
	} else {
		return not
	}
}

func FormatPostBody(post PostData) string {
	output := color.HiBlackString("Mime Type: ") + TypeColor(post.MimeType)
	if strings.Contains(post.MimeType, "application/json") || IsValidJson(post.Text) {
		var i interface{}
		err := json.Unmarshal([]byte(post.Text), &i)
		if err == nil {
			if !strings.Contains(post.MimeType, "application/json") {
				output += color.HiBlackString(" (inferred application/json)")
			}
			output += "\n"

			formatter := colorjson.NewFormatter()
			formatter.Indent = 2
			processed, err := formatter.Marshal(i)
			if err == nil {
				output += string(processed)
				return output
			} else {
				slog.Error("Failed to color the json", "error", err)
			}
		} else {
			slog.Error("Failed to unmarshall the json into this type", "error", err)
		}
	} else {
		output += "\n" + post.Text
	}

	if len(post.Params) > 0 {
		output += color.YellowString("\nParameters: \n")
		for _, param := range post.Params {
			output += "  " + param.Name + " = "
			if param.Value != nil {
				output += TypeColor(*param.Value)
			} else {
				output += "[no value]"
			}
			if param.ContentType != nil {
				output += color.HiBlackString("\n    Content Type: ") + TypeColor(*param.ContentType)
			}
			if param.ContentType != nil {
				output += color.HiBlackString("\n    File Name: ") + TypeColor(*param.FileName)
			}
			if param.Comment != nil {
				output += color.HiBlackString("\n    Comment: ") + TypeColor(*param.Comment)
			}
		}
	}

	return output
}
func FormatContent(post Content) string {
	headers := color.HiBlackString("Size: ") + TypeColor(strconv.Itoa(post.Size)) + "\n"

	if post.Encoding != nil {
		headers += color.HiBlackString("Encoding: ") + TypeColor(*post.Encoding) + "\n"
	}
	if post.Compression != nil {
		headers += color.HiBlackString("Compression: ") + TypeColor(strconv.Itoa(*post.Compression)) + "\n"
	}

	if post.Text != nil && (strings.Contains(post.MimeType, "application/json") || IsValidJson(*post.Text)) {
		var i interface{}
		err := json.Unmarshal([]byte(*post.Text), &i)
		if err == nil {
			output := color.HiBlackString("Mime Type: ") + TypeColor(post.MimeType)
			if !strings.Contains(post.MimeType, "application/json") {
				output += color.HiBlackString(" (inferred application/json)")
			}
			output += "\n"
			output += headers

			formatter := colorjson.NewFormatter()
			formatter.Indent = 2
			processed, err := formatter.Marshal(i)
			if err == nil {
				output += string(processed)
				return output
			} else {
				slog.Error("Failed to color the json", "error", err)
			}
		} else {
			slog.Error("Failed to unmarshall the json into this type", "error", err)
		}
	}

	if post.Text != nil {
		return headers + *post.Text
	} else {
		return headers + "[no text]"
	}
}

func IsValidJson(v string) bool {
	var i interface{}
	err := json.Unmarshal([]byte(v[:]), &i)
	return err == nil
}

func TypeColor(s string) string {
	_, err := strconv.ParseFloat(s, 64)
	isNumber := err == nil

	if isNumber {
		return color.YellowString(s)
	}
	if strings.ToLower(s) == "false" && strings.ToLower(s) == "true" {
		return color.MagentaString(s)
	}

	return color.CyanString(s)
}

func Indent(v string, spaces int) string {
	lines := strings.Split(v, "\n")
	indented := make([]string, len(lines))
	for i := 0; i < len(indented); i++ {
		indented[i] = strings.Repeat(" ", spaces) + lines[i]
	}
	return strings.Join(indented, "\n")
}

func FormatEntry(entry Entry) string {
	result := color.YellowString(strings.ToLower(entry.Request.HttpVersion)+" "+entry.Request.Method) + " " + entry.Request.Url
	if CLI.IncludeHeaders != nil && *CLI.IncludeHeaders {
		result += color.YellowString("\n  Request Headers:")
		for _, header := range entry.Request.Headers {
			if strings.ToLower(header.Name) == "cookie" {
				continue
			}
			result += "\n    " + color.HiBlackString(header.Name) + " = " + TypeColor(header.Value)
			if header.Comment != nil {
				result += " (" + *header.Comment + ")"
			}
		}
	}
	if CLI.IncludeCookies != nil && *CLI.IncludeCookies && len(entry.Request.Cookies) > 0 {
		result += color.YellowString("\n  Request Cookies:")
		for _, header := range entry.Request.Cookies {
			result += "\n    " + color.HiBlackString(header.Name) + " = " + TypeColor(header.Value)
			if header.Comment != nil {
				result += " (" + *header.Comment + ")"
			}
		}
	}
	if CLI.IncludeRequestBody != nil && *CLI.IncludeRequestBody && entry.Request.PostData != nil {
		if entry.Request.BodySize == 0 {
			result += color.YellowString("\n  Request Body:\n    ") + "[no content]"
		} else {
			result += color.YellowString("\n  Request Body:\n") + Indent(FormatPostBody(*entry.Request.PostData), 4)
		}
	}

	if CLI.IncludeHeaders != nil && *CLI.IncludeHeaders {
		result += color.YellowString("\n  Response Headers:")
		for _, header := range entry.Response.Headers {
			if strings.ToLower(header.Name) == "cookie" {
				continue
			}
			result += "\n    " + color.HiBlackString(header.Name) + " = " + TypeColor(header.Value)
			if header.Comment != nil {
				result += " (" + *header.Comment + ")"
			}
		}
	}
	if CLI.IncludeCookies != nil && *CLI.IncludeCookies && len(entry.Response.Cookies) > 0 {
		result += color.YellowString("\n  Response Cookies:")
		for _, header := range entry.Response.Cookies {
			result += "\n    " + color.HiBlackString(header.Name) + " = " + TypeColor(header.Value)
			if header.Comment != nil {
				result += " (" + *header.Comment + ")"
			}
		}
	}
	if CLI.IncludeResponseBody != nil && *CLI.IncludeResponseBody && entry.Response.Content != nil {
		if (*entry.Response.Content).Size == 0 {
			result += color.YellowString("\n  Response Body:\n    ") + "[no content]"
		} else {
			result += color.YellowString("\n  Response Body:\n") + Indent(FormatContent(*entry.Response.Content), 4)
		}
	}
	if CLI.IncludeTimings != nil && *CLI.IncludeTimings {
		result += color.YellowString("\n  Timings:    ")
		if entry.Timings.Dns != nil && *entry.Timings.Dns >= 0 {
			result += color.HiBlackString("\n        DNS: ") + TypeColor(strconv.Itoa(*entry.Timings.Dns))
		}
		if entry.Timings.Connect != nil && *entry.Timings.Connect >= 0 {
			result += color.HiBlackString("\n    Connect: ") + TypeColor(strconv.Itoa(*entry.Timings.Connect))
		}
		result += color.HiBlackString("\n       Send: ") + TypeColor(strconv.Itoa(entry.Timings.Send))
		result += color.HiBlackString("\n       Wait: ") + TypeColor(strconv.Itoa(entry.Timings.Wait))
		result += color.HiBlackString("\n    Receive: ") + TypeColor(strconv.Itoa(entry.Timings.Receive))
		if entry.Timings.Ssl != nil && *entry.Timings.Ssl >= 0 {
			result += color.HiBlackString("\n        SSL: ") + TypeColor(strconv.Itoa(*entry.Timings.Ssl))
		}
		if entry.Timings.Comment != nil {
			result += color.HiBlackString("\n    Comment: ") + *entry.Comment
		}
	}

	return result
}

func main() {
	kong.Parse(&CLI,
		kong.Name("harv"),
		kong.Description("A simple command line HAR file viewer"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))

	//spew.Dump(CLI)
	content, err := os.ReadFile(CLI.File)
	if err != nil {
		panic(err)
		return
	}

	var har HarFile
	json.Unmarshal(content, &har)

	validEntries := Filter(har.Log.Entries, IsEntryValid)
	//spew.Dump(validEntries)
	//fmt.Printf("Found %+v valid entries\n", len(validEntries))
	for _, entry := range validEntries {
		println(FormatEntry(entry))
	}
}
