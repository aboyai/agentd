// cmd/server/visualizer.go
package main

import (
	"agentd/router"
	"fmt"
	"net/http"
)

func StartVisualizerServer() {
	http.HandleFunc("/visualize", func(w http.ResponseWriter, r *http.Request) {
		dagStr := r.URL.Query().Get("dag")
		if dagStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Missing ?dag=plan://... parameter")
			return
		}

		dag := router.NewDAGFromPlan(dagStr)
		diagram := dag.MermaidString()

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>AgentD DAG Viewer</title>
  <script type="module">
    import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs';
    mermaid.initialize({ startOnLoad: true, theme: 'dark' });
  </script>
</head>
<body style="background:black; color:#00d8ff; font-family:sans-serif; text-align:center; padding:2em;">
  <h2>AgentD DAG Visualization</h2>
  <div class="mermaid">
%s
  </div>
</body>
</html>`, diagram)
	})

	fmt.Println("🌐 AgentD Visualizer at http://localhost:8080/visualize")
	http.ListenAndServe(":8080", nil)
}
