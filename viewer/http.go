package viewer

import (
	"log"
	"net/http"
	"html/template"
	"github.com/ld86/godht/node"
)

type HttpViewer struct {
	n *node.Node
}

type BucketInformation struct {
	Index int
	Size int
	Ids [][20]byte
}

type RootData struct {
	NodeId [20]byte
	Buckets [160]BucketInformation
}

func NewHttpViewer(n *node.Node) *HttpViewer {
	return &HttpViewer{n: n}
}

const rootTemplate = `
<html>
	<head>
	</head>
	<body>
		<table>
			<tr>
				<td>
					My Id
				</td>
				<td>
					{{.NodeId}}
				</td>
			</tr>
			{{with .Buckets}}
				{{range .}}
					{{if .Size}}
					<tr>
						<td>
							{{.Index}}
						</td>
						<td>
							{{.Ids}}
						</td>
					</tr>
					{{end}}
				{{end}}
			{{end}}
		</table>
	</body>
</html>`

func (viewer *HttpViewer) rootHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t := template.New("root")
		t, err := t.Parse(rootTemplate)
		if err != nil {
			log.Fatalf("Cannot parse root template")
		}

		data := RootData{NodeId: viewer.n.Id()}

		for i := 0; i < 160; i++ {
			bucket := viewer.n.Buckets().GetBucket(159 - i)

			if bucket.Len() == 0 {
				continue
			}

			var bucketInformation BucketInformation
			bucketInformation.Index = 159 - i;
			bucketInformation.Size = bucket.Len()
			bucketInformation.Ids = make([][20]byte, 0)

			for it := bucket.Front(); it != nil; it = it.Next() {
				bucketInformation.Ids = append(bucketInformation.Ids, it.Value.([20]byte))
			}

			data.Buckets[i] = bucketInformation
		}


		t.Execute(w, data)
	}
}

func (viewer *HttpViewer) Serve() {
	http.HandleFunc("/", viewer.rootHandler())
	http.ListenAndServe(":8080", nil)
}
