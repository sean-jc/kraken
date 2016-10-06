package kraken

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// A Report represents the state a Reporter uses to write out its reports.
type Report interface {
	// Add adds a given *Result to a Report.
	Add(*Result)
}

// Closer wraps the optional Report Close method.
type Closer interface {
	// Close permantently closes a Report, running any necessary book keeping.
	Close()
}

// A Reporter function writes out reports to the given io.Writer or returns an
// error in case of failure.
type Reporter func(io.Writer) error

// Report is a convenience method wrapping the Reporter function type.
func (rep Reporter) Report(w io.Writer) error { return rep(w) }

// NewHistogramReporter returns a Reporter that writes out a Histogram as
// aligned, formatted text.
func NewHistogramReporter(h *Histogram) Reporter {
	return func(w io.Writer) (err error) {
		tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', tabwriter.StripEscape)
		for _, t := range h.Tables {
			if _, err = fmt.Fprintf(tw, "\n%s\n\nBucket\t\t#\t%%\tHistogram\n", t.Name); err != nil {
				return err
			}
			for i, count := range t.Counts {
				ratio := float64(count) / float64(t.Total)
				lo, hi := h.Buckets.Nth(i)
				pad := strings.Repeat("#", int(ratio*75))
				_, err = fmt.Fprintf(tw, "[%s,\t%s]\t%d\t%.2f%%\t%s\n", lo, hi, count, ratio*100, pad)
				if err != nil {
					return nil
				}
			}

		}

		return tw.Flush()
	}
}

// NewTextReporter returns a Reporter that writes out Metrics as aligned,
// formatted text.
func NewTextReporter(m *Metrics) Reporter {
	const hstr = "Hits\t[total, rate]\t%d, %.2f\n" +
		"Success\t[ratio]\t%.2f%%\n" +
		"Duration\t[total, attack, wait]\t%s, %s, %s\n"
	const lstr = "%s\t[mean, 50, 95, 99, max]\t%s, %s, %s, %s, %s\n"

	return func(w io.Writer) error {
		tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', tabwriter.StripEscape)
		_, err := fmt.Fprintf(tw, hstr,
			m.Hits, m.Rate,
			m.Success*100,
			m.Duration+m.Wait, m.Duration, m.Wait,
		)
		if err != nil {
			return err
		}

		for _, l := range m.Latencies {
			if _, err = fmt.Fprintf(tw, lstr, l.Name, l.Mean, l.P50, l.P95, l.P99, l.Max); err != nil {
				return err
			}
		}

		if len(m.Errors) > 0 {
			if _, err = fmt.Fprintln(tw, "\nError Set:"); err != nil {
				return err
			}

			for _, e := range m.Errors {
				if _, err = fmt.Fprintln(tw, e); err != nil {
					return err
				}
			}
		}
		return tw.Flush()
	}
}

// NewJSONReporter returns a Reporter that writes out Metrics as JSON.
func NewJSONReporter(m *Metrics) Reporter {
	return func(w io.Writer) error {
		return json.NewEncoder(w).Encode(m)
	}
}

// // NewPlotReporter returns a Reporter that writes a self-contained
// // HTML page with an interactive plot of the latencies of Requests, built with
// // http://dygraphs.com/
// func NewPlotReporter(title string, rs *Results) Reporter {
// 	return func(w io.Writer) (err error) {
// 		_, err = fmt.Fprintf(w, plotsTemplateHead, title, asset(dygraphs), asset(html2canvas))
// 		if err != nil {
// 			return err
// 		}

// 		buf := make([]byte, 0, 128)
// 		for i, result := range *rs {
// 			buf = append(buf, '[')
// 			buf = append(buf, strconv.FormatFloat(
// 				result.Timestamp.Sub((*rs)[0].Timestamp).Seconds(), 'f', -1, 32)...)
// 			buf = append(buf, ',')

// 			latency := strconv.FormatFloat(result.Latency.Seconds()*1000, 'f', -1, 32)
// 			if result.Error == "" {
// 				buf = append(buf, "NaN,"...)
// 				buf = append(buf, latency...)
// 				buf = append(buf, ']', ',')
// 			} else {
// 				buf = append(buf, latency...)
// 				buf = append(buf, ",NaN],"...)
// 			}

// 			if i == len(*rs)-1 {
// 				buf = buf[:len(buf)-1]
// 			}

// 			if _, err = w.Write(buf); err != nil {
// 				return err
// 			}

// 			buf = buf[:0]
// 		}

// 		_, err = fmt.Fprintf(w, plotsTemplateTail, title)
// 		return err
// 	}
// }

// const (
// 	plotsTemplateHead = `<!doctype html>
// <html>
// <head>
//   <title>%s</title>
// </head>
// <body>
//   <div id="latencies" style="font-family: Courier; width: 100%%; height: 600px"></div>
//   <button id="download">Download as PNG</button>
//   <script>%s</script>
//   <script>%s</script>
//   <script>
//   new Dygraph(
//     document.getElementById("latencies"),
//     [`
// 	plotsTemplateTail = `],
//     {
//       title: '%s',
//       labels: ['Seconds', 'ERR', 'OK'],
//       ylabel: 'Latency (ms)',
//       xlabel: 'Seconds elapsed',
//       showRoller: true,
//       colors: ['#FA7878', '#8AE234'],
//       legend: 'always',
//       logscale: true,
//       strokeWidth: 1.3
//     }
//   );
//   document.getElementById("download").addEventListener("click", function(e) {
//     html2canvas(document.body, {background: "#fff"}).then(function(canvas) {
//       var url = canvas.toDataURL('image/png').replace(/^data:image\/[^;]/, 'data:application/octet-stream');
//       var a = document.createElement("a");
//       a.setAttribute("download", "vegeta-plot.png");
//       a.setAttribute("href", url);
//       a.click();
//     });
//   });
//   </script>
// </body>
// </html>`
// )
