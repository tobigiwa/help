package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/a-h/templ"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func main() {

	server := &http.Server{
		Addr:    ":8080",
		Handler: Routes(),
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func(done chan os.Signal) {
		<-done
		close(done)

		if err := server.Shutdown(context.TODO()); err != nil {
			log.Fatalf("Graceful server shutdown Failed:%+v\n", err)
		}
	}(done)

	fmt.Println("Server is running on port http://127.0.0.1:8080/screentime")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln("Server error:", err)
	}
	fmt.Println()
	fmt.Println("SERVER STOPPED GRACEFULLY")
}

func Routes() *http.ServeMux {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets", fs))

	mux.HandleFunc("GET /screentime", ScreenTimePageHandler)

	return mux
}

func ScreenTimePageHandler(w http.ResponseWriter, r *http.Request) {
	if err := ScreenTimePage().Render(context.TODO(), w); err != nil {
		return
	}
}

type Renderable interface {
	Render(w io.Writer) error
}

func ConvertChartToTemplComponent(chart Renderable) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return chart.Render(w)
	})
}

func generateBarItems() []opts.BarData {
	items := make([]opts.BarData, 0)
	for i := 0; i < 7; i++ {
		items = append(items, opts.BarData{Value: rand.Intn(300)})
	}
	return items
}

func createBarChart() *charts.Bar {
	const actionsWithEchartInstance = `
 const myChart = %MY_ECHARTS%;
 window.onresize = function ()
     {
         myChart.resize();
     };
`
	bar := charts.NewBar()
	bar.AddJSFuncs(opts.FuncOpts(actionsWithEchartInstance))
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Bar chart",
		Subtitle: "That works well with templ",
	}))
	bar.SetXAxis([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		AddSeries("Category A", generateBarItems()).
		AddSeries("Category B", generateBarItems())
	return bar
}
