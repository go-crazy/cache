package sessions_test

import (
	"testing"
	"encoding/json"
	"github.com/go-crazy/cache"
	Gin "github.com/gin-gonic/gin"
	"github.com/gavv/httpexpect"
)

func TestSessions(t *testing.T) {
	app := Gin.Default()

	sess := sessions.New(sessions.Config{Cookie: "mycustomsessionid"})
	testSessions(t, sess, app)
	
	// app.Run(":8099")
}

const (
	testEnableSubdomain = false
)

func testSessions(t *testing.T, sess *sessions.Sessions, app *Gin.Engine ) {
	values := map[string]interface{}{
		"Name":   "iris",
		"Months": "4",
		"Secret": "dsads£2132215£%%Ssdsa",
	}

	writeValues := func(ctx *Gin.Context) {
		s := sess.Start(ctx)
		sessValues := s.GetAll()

		ctx.JSON(200,sessValues)
	}

	if testEnableSubdomain {
		app.Group("subdomain.").GET("/get", func(ctx *Gin.Context) {
			writeValues(ctx)
		})
	}

	app.POST("/set", func(ctx *Gin.Context) {
		s := sess.Start(ctx)
		vals := make(map[string]interface{}, 0)
		
		decoder := json.NewDecoder(ctx.Request.Body)
		
		err := decoder.Decode(&vals)
		if err != nil {
			panic(err)
		}
		defer ctx.Request.Body.Close()

		for k, v := range vals {
			s.Set(k, v)
		}
	})

	app.GET("/get", func(ctx *Gin.Context) {
		writeValues(ctx)
	})

	app.GET("/clear", func(ctx *Gin.Context) {
		sess.Start(ctx).Clear()
		writeValues(ctx)
	})

	app.GET("/destroy", func(ctx *Gin.Context) {
		sess.Destroy(ctx)
		writeValues(ctx)
		// the cookie and all values should be empty
	})

	// request cookie should be empty
	app.GET("/after_destroy", func(ctx *Gin.Context) {
	})

	app.GET("/multi_start_set_get", func(ctx *Gin.Context) {
		s := sess.Start(ctx)
		s.Set("key", "value")
		ctx.Next()
	}, func(ctx *Gin.Context) {
		s := sess.Start(ctx)
		ctx.String(200,s.GetString("key"))
	})

	go app.Run(":8099")
	e := httpexpect.New(t, "http://127.0.0.1:8099")

	e.POST("/set").WithJSON(values).Expect().Status(200).Cookies().NotEmpty()
	e.GET("/get").Expect().Status(200).JSON().Object().Equal(values)
	if testEnableSubdomain {
		es := httpexpect.New(t, "http://127.0.0.1:8099")
		es.Request("GET", "/get").Expect().Status(200).JSON().Object().Equal(values)
	}

	// test destroy which also clears first
	d := e.GET("/destroy").Expect().Status(200)
	d.JSON().Object().Empty()
	// 	This removed: d.Cookies().Empty(). Reason:
	// httpexpect counts the cookies setted or deleted at the response time, but cookie is not removed, to be really removed needs to SetExpire(now-1second) so,
	// test if the cookies removed on the next request, like the browser's behavior.
	e.GET("/after_destroy").Expect().Status(200).Cookies().Empty()
	// set and clear again
	// e.POST("/set").WithJSON(values).Expect().Status(200).Cookies().NotEmpty()
	e.GET("/clear").Expect().Status(200).JSON().Object().Empty()

	// test start on the same request but more than one times

	e.GET("/multi_start_set_get").Expect().Status(200).Body().Equal("value")
}

func TestFlashMessages(t *testing.T) {
	app := Gin.Default()

	sess := sessions.New(sessions.Config{Cookie: "mycustomsessionid"})

	valueSingleKey := "Name"
	valueSingleValue := "iris-sessions"

	values := map[string]interface{}{
		valueSingleKey: valueSingleValue,
		"Days":         "1",
		"Secret":       "dsads£2132215£%%Ssdsa",
	}

	writeValues := func(ctx *Gin.Context, values map[string]interface{}) {
	

		ctx.JSON(200,values)
	}

	// ctx.JSON(200,values)
	// writeValues(ctx)
	// writeValues := func(ctx *Gin.Context, values map[string]interface{}) error {
		
	// }

	app.POST("/set", func(ctx *Gin.Context) {
		vals := make(map[string]interface{}, 0)

		decoder := json.NewDecoder(ctx.Request.Body)
		
		err := decoder.Decode(&vals)
		if err != nil {
			panic(err)
		}
		defer ctx.Request.Body.Close()
		// if err := ctx.ReadJSON(&vals); err != nil {
		// 	t.Fatalf("Cannot readjson. Trace %s", err.Error())
		// }
		s := sess.Start(ctx)
		for k, v := range vals {
			s.SetFlash(k, v)
		}

		ctx.Status(200)
	})

	writeFlashValues := func(ctx *Gin.Context) {
		s := sess.Start(ctx)

		flashes := s.GetFlashes()
		writeValues(ctx, flashes)
		// if err := writeValues(ctx, flashes); err != nil {
		// 	t.Fatalf("While serialize the flash values: %s", err.Error())
		// }
	}

	app.GET("/get_single", func(ctx *Gin.Context) {
		s := sess.Start(ctx)
		flashMsgString := s.GetFlashString(valueSingleKey)
		ctx.String(200,flashMsgString)
	})

	app.GET("/get", func(ctx *Gin.Context) {
		writeFlashValues(ctx)
	})

	app.GET("/clear", func(ctx *Gin.Context) {
		s := sess.Start(ctx)
		s.ClearFlashes()
		writeFlashValues(ctx)
	})

	app.GET("/destroy", func(ctx *Gin.Context) {
		sess.Destroy(ctx)
		writeFlashValues(ctx)
		ctx.Status(200)
		// the cookie and all values should be empty
	})
	go app.Run(":8098")
	// request cookie should be empty
	app.GET("/after_destroy", func(ctx *Gin.Context) {
		ctx.Status(200)
	})
	
	e := httpexpect.New(t, "http://127.0.0.1:8098")

	e.POST("/set").WithJSON(values).Expect().Status(200).Cookies().NotEmpty()
	// get all
	e.GET("/get").Expect().Status(200).JSON().Object().Equal(values)
	// get the same flash on other request should return nothing because the flash message is removed after fetch once
	e.GET("/get").Expect().Status(200).JSON().Object().Empty()
	// test destroy which also clears first
	d := e.GET("/destroy").Expect().Status(200)
	d.JSON().Object().Empty()
	e.GET("/after_destroy").Expect().Status(200).Cookies().Empty()
	// set and clear again
	// e.POST("/set").WithJSON(values).Expect().Status(200).Cookies().NotEmpty()
	e.GET("/clear").Expect().Status(200).JSON().Object().Empty()

	// set again in order to take the single one ( we don't test Cookies.NotEmpty because httpexpect default conf reads that from the request-only)
	e.POST("/set").WithJSON(values).Expect().Status(200)
	e.GET("/get_single").Expect().Status(200).Body().Equal(valueSingleValue)

}
