package main

import (
	"log"
	"net/http"
	"os"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/yeouchien/go-graphql-subscription-example/go/graph"
	"github.com/yeouchien/go-graphql-subscription-example/go/wshandler"
)

func main() {
	r, err := graph.NewResolver(os.Getenv("OPENTSDB_HOST"))
	if err != nil {
		log.Fatalf("error getting resolver: %v", err)
	}

	// init graphQL schema
	s, err := graphql.ParseSchema(graph.Schema, r)
	if err != nil {
		log.Fatalf("error getting schema: %v", err)
	}

	graphqlHandler := &relay.Handler{
		Schema: s,
	}

	http.HandleFunc("/graphiql", func(w http.ResponseWriter, r *http.Request) {
		w.Write(graphiqlHTML)
	})
	http.HandleFunc("/graphql", allowCors(graphqlHandler))
	http.Handle("/subscriptions", wshandler.NewWSHandler(s))

	addr := ":" + os.Getenv("PORT")
	if addr == ":" {
		addr = ":5000"
	}

	log.Println("Listening at", addr, "...")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

var graphiqlHTML = []byte(`
<!DOCTYPE html>
<html>
    <head>
        <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.10.2/graphiql.css" />
        <script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/1.1.0/fetch.min.js"></script>
        <script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react.min.js"></script>
        <script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react-dom.min.js"></script>
        <script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.10.2/graphiql.js"></script>
    </head>
    <body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
        <div id="graphiql" style="height: 100vh;">Loading...</div>
        <script>
            function graphQLFetcher(graphQLParams) {
                return fetch("/graphql", {
                    method: "post",
                    body: JSON.stringify(graphQLParams),
                    credentials: "include",
                }).then(function (response) {
                    return response.text();
                }).then(function (responseBody) {
                    try {
                        return JSON.parse(responseBody);
                    } catch (error) {
                        return responseBody;
                    }
                });
            }
            ReactDOM.render(
                React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
                document.getElementById("graphiql")
            );
        </script>
    </body>
</html>
`)

func allowCors(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if r.Method == "OPTIONS" {
			return
		}
		handler.ServeHTTP(w, r)
	}
}
