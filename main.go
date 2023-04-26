package main

import (
	"EIP/model"
	"EIP/service"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/websocket"
	"log"
	"strconv"
	"strings"
	"time"
)

var unixSocket string
var klipperPath string
var messageChan = make(chan string)
var conns []*websocket.Conn

func main() {

	ws := websocket.New(websocket.DefaultGorillaUpgrader, websocket.Events{
		websocket.OnNativeMessage: func(nsConn *websocket.NSConn, msg websocket.Message) error {
			return nil
		},
	})

	ws.OnConnect = func(c *websocket.Conn) error {
		conns = append(conns, c)
		for {
			message := <-messageChan
			mg := websocket.Message{
				Body:     []byte(message),
				IsNative: true,
			}
			for _, conn := range conns {
				go conn.Write(mg)
			}
		}
		return nil
	}

	app := iris.New()
	tmpl := iris.HTML("./views", ".html")
	// Set custom delimeters.
	tmpl.Delims("{{", "}}")
	//tmpl.Reload(true)

	app.RegisterView(tmpl)

	app.Get("/", home)
	app.Get("/report/{ctime:int64}", report)

	app.Get("/ws", websocket.Handler(ws))
	app.Post("/config/socket", socketConfig)
	app.Post("/config/path", pathConfig)
	app.Post("/report/rename", reName)
	app.Post("/report/del", del)
	app.Post("/newCalibrate", calibrate)

	app.HandleDir("/static", iris.Dir("./static"))

	app.Listen(":8080")
}

func home(ctx iris.Context) {

	records := model.GetRecords()

	ctx.ViewData("records", records)
	ctx.ViewData("cTime", records[len(records)-1].Time)

	if err := ctx.View("index.html"); err != nil {
		ctx.HTML("<h3>%s</h3>", err.Error())

	}
}

func report(ctx iris.Context) {
	cTime, err := ctx.Params().GetInt64("ctime")
	if err != nil {
		ctx.StatusCode(500)
	}
	// query
	records := model.GetRecords()

	new := true
	for i := range records {
		if records[i].Time == cTime {
			new = false
			break
		}
	}
	ctx.ViewData("records", records)
	ctx.ViewData("cTime", cTime)

	ctx.ViewData("cTimeStr", time.Unix(cTime, 0).Format("2006-01-02 15:04"))
	ctx.ViewData("new", new)
	if err := ctx.View("index.html"); err != nil {
		ctx.HTML("<h3>%s</h3>", err.Error())

	}
}

func socketConfig(ctx iris.Context) {
	unixSocket = ctx.PostValue("input")
	log.Println(unixSocket)
	model.UpdateSocket(unixSocket)
	ctx.JSON(iris.Map{"message": "OK"})
}
func pathConfig(ctx iris.Context) {
	klipperPath = ctx.PostValue("input")

	klipperPath = strings.TrimRight(klipperPath, "/")
	if !strings.HasSuffix(klipperPath, "/") {
		klipperPath += "/"
	}

	log.Println(klipperPath)
	model.UpdatePath(klipperPath)
	ctx.JSON(iris.Map{"message": "OK"})
}
func reName(ctx iris.Context) {
	newName := ctx.PostValue("name")
	cTime, _ := ctx.PostValueInt64("ctime")
	model.UpdateName(newName, cTime)
	//ctx.Redirect("/record/" + strconv.FormatInt(cTime, 10))
	ctx.JSON(iris.Map{"message": "OK"})

}

func del(ctx iris.Context) {
	cTime, _ := ctx.PostValueInt64("ctime")
	model.DelRecord(cTime)
	ctx.Redirect("/")
}

func calibrate(ctx iris.Context) {
	cTime := time.Now().Unix()
	service.NewCalibrate(unixSocket, klipperPath, messageChan, cTime)
	ctx.Redirect("/report/"+strconv.FormatInt(cTime, 10), iris.StatusFound)

}
