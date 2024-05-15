package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	chartrender "github.com/go-echarts/go-echarts/v2/render"
)

var htmlSnippet *template.Template

func main() {
	pie := charts.NewPie()

	pie.Renderer = newSnippetRenderer(pie, pie.Validate)

	pieData := []opts.PieData{
		{Name: "Dead Cases", Value: 123},
		{Name: "Recovered Cases", Value: 456},
		{Name: "Active Cases", Value: 789},
	}

	// put data into chart
	pie.AddSeries("Case Distribution", pieData).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{Show: true, Formatter: "{b}: {c}"}),
	)

	htmlSnippet = renderToHtml(pie)

	server := &http.Server{
		Addr:    ":8081",
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

	fmt.Println("Server is running on port http://127.0.0.1:8081/screentime")
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

type Renderer interface {
	Render(w io.Writer) error
}

type snippetRenderer struct {
	c      interface{}
	before []func()
}

func renderToHtml(c interface{}) *template.Template {
	var buf bytes.Buffer
	r := c.(chartrender.Renderer)
	err := r.Render(&buf)
	if err != nil {
		log.Printf("Failed to render chart: %s", err)
		return nil
	}

	return template.New(buf.String())
}

func newSnippetRenderer(c interface{}, before ...func()) chartrender.Renderer {
	return &snippetRenderer{c: c, before: before}
}

func (r *snippetRenderer) Render(w io.Writer) error {
	const tplName = "chart"
	for _, fn := range r.before {
		fn()
	}

	tpl := template.
		Must(template.New(tplName).
			Funcs(template.FuncMap{
				"safeJS": func(s interface{}) template.JS {
					return template.JS(fmt.Sprint(s))
				},
			}).
			Parse(baseTpl),
		)

	err := tpl.ExecuteTemplate(w, tplName, r.c)
	return err
}

var baseTpl = `
<div class="container">
    <div class="item" id="{{ .ChartID }}" style="width:{{ .Initialization.Width }};height:{{ .Initialization.Height }};"></div>
</div>
{{- range .JSAssets.Values }}
   <script src="{{ . }}"></script>
{{- end }}
<script type="text/javascript">
    "use strict";
    let goecharts_{{ .ChartID | safeJS }} = echarts.init(document.getElementById('{{ .ChartID | safeJS }}'), "{{ .Theme }}");
    let option_{{ .ChartID | safeJS }} = {{ .JSON }};
    goecharts_{{ .ChartID | safeJS }}.setOption(option_{{ .ChartID | safeJS }});
    {{- range .JSFunctions.Fns }}
    {{ . | safeJS }}
    {{- end }}
</script>
`























// func createBarChart() *charts.Bar {
// 	bar := charts.NewBar()
// 	bar.SetGlobalOptions(
// 		charts.WithTitleOpts(opts.Title{
// 			Title:    "title and legend options",
// 			Subtitle: "go-echarts is an awesome chart library written in Golang",
// 			Link:     "https://github.com/go-echarts/go-echarts",
// 			Right:    "40%",
// 		}),
// 		charts.WithLegendOpts(opts.Legend{Right: "80%"}),
// 	)
// 	// bar.AddJSFuncs(opts.FuncOpts(actionsWithEchartInstance))
// 	bar.SetXAxis(weeks).
// 		AddSeries("Category A", generateBarItems()).
// 		AddSeries("Category B", generateBarItems())
// 	return bar
// }

// var (
// 	itemCnt = 7
// 	weeks   = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
// )

// type Renderable interface {
// 	Render(w io.Writer) error
// }

// func ConvertChartToTemplComponent(chart Renderable) templ.Component {
// 	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
// 		return chart.Render(w)
// 	})
// }

// func generateBarItems() []opts.BarData {
// 	items := make([]opts.BarData, 0)
// 	for i := 0; i < 7; i++ {
// 		items = append(items, opts.BarData{Value: rand.Intn(300)})
// 	}
// 	return items
// }
