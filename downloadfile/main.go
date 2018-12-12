package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	port string
)

func init() {
	p := flag.String("p", "36245", "port")
	flag.Parse()

	port = *p
}

func main() {
	e := echo.New()

	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize:         2,
		DisableStackAll:   true,
		DisablePrintStack: true,
	}))
	e.Use(middleware.RemoveTrailingSlash())

	e.GET("/download/:file", func(c echo.Context) error {
		f := c.Param("file")

		file, e := os.Open("./" + f)
		if e != nil {
			fmt.Println(e.Error())
			return c.String(http.StatusBadRequest, e.Error())
		}
		defer file.Close()

		info, e := file.Stat()
		if e != nil {
			fmt.Println(e.Error())
			return c.String(http.StatusBadRequest, e.Error())
		} else if info.IsDir() {
			return c.String(http.StatusBadRequest, fmt.Sprintf("%s is dir", info.Name()))
		}
		fmt.Println("downloading", f)
		defer func() {
			fmt.Println("download over", f)
		}()

		c.Response().Header().Set("Expires", "0")
		c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", f))
		c.Response().Header().Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
		c.Response().Header().Set("Content-Transfer-Encoding", "binary")
		c.Response().Header().Set("Pragma", "public")
		c.Response().Header().Set(echo.HeaderContentLength, fmt.Sprintf("%v", info.Size()))
		return c.Stream(http.StatusOK, echo.MIMEOctetStream, file)
	})

	e.GET("/list", func(c echo.Context) error {
		// files, e := filepath.Glob("*")
		files, e := ioutil.ReadDir(".")
		if e != nil {
			return c.String(http.StatusBadRequest, e.Error())
		}
		names := make([]string, len(files))
		for index, item := range files {
			if item.IsDir() {
				names[index] = item.Name() + "/"
			} else {
				names[index] = item.Name()
			}
		}
		return c.String(http.StatusOK, strings.Join(names, "\n"))
	})

	e.Start("0.0.0.0:" + port)
}
