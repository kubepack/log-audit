package main

import (
	"fmt"
	"sync"
	"net/http"
	"html"
	"io/ioutil"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/appscode/go/log"
	"os"
	"path/filepath"
	"encoding/json"
	"k8s.io/apiserver/pkg/apis/audit/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AppName = "log-audit"
)
var mux sync.Mutex

func main() {
	fmt.Println("Server Started...")
	routine := 0
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events" {
			http.NotFound(w, r)
			return
		}
		fmt.Println("hello request")
		fmt.Fprintf(w, "Hello %q", html.EscapeString(r.URL.Path))
		resp, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(string(resp))
		eventList := &v1beta1.EventList{}
		err = json.Unmarshal(resp, eventList)
		if err != nil {
			log.Fatalln(err)
		}
		routine += 1
		go ProcessEvents(eventList, routine)
		if err != nil {
			fmt.Println(err)
			log.Fatalln(err)
		}
	})

	http.HandleFunc("/get-logs", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/get-logs" {
			return
		}
		fmt.Println("hello request222")
		fmt.Fprintf(writer, "Hello %q", html.EscapeString(request.URL.Path))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ProcessEvents(list *v1beta1.EventList, routine int) error {
	fmt.Printf("Routine Number %d\n", routine)
	fmt.Printf("Hello %s\n", list.Kind)
	if list == nil {
		fmt.Println("Nil")
		return fmt.Errorf("%s", "Empty event list")
	}

	mapToGitCommitHash := map[string]*v1beta1.EventList{}
	eventList := &v1beta1.EventList{}
	var events []*v1beta1.Event
	for _, val := range list.Items {
		fmt.Println("-----------------------")
		fmt.Println(routine)
		_, err := json.MarshalIndent(val, "", "  ")
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(val.ObjectRef)
		if val.ResponseObject != nil {
			fmt.Println("********************")

			type Item struct {
				metav1.TypeMeta   `json:",inline"`
				metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
			}

			item := &Item{}

			err := json.Unmarshal(val.ResponseObject.Raw, item)
			if err != nil {
				return err
			}

			gitCommitHash, ok := item.Annotations["git-commit-hash"]
			if ok {
				fmt.Println("Git Commit Hash Present....", gitCommitHash)
				evByte, err := json.Marshal(val)
				if err != nil {
					fmt.Println("Error during marshalling event")
				}

				eventTmp := &v1beta1.Event{}
				err = json.Unmarshal(evByte, eventTmp)
				if err != nil {
					fmt.Println("Error during Unmarshall event byte")
				}
				events = append(events, eventTmp)

				if _, ok = mapToGitCommitHash[gitCommitHash]; !ok {
					mapToGitCommitHash[gitCommitHash] = &v1beta1.EventList{}
				}
				eventList = mapToGitCommitHash[gitCommitHash]
				eventList.Items = append(eventList.Items, *eventTmp)
			}
		}
	}
	mux.Lock()
	defer mux.Unlock()
	db, err := OpenGolevelDB()
	defer db.Close()
	if err != nil {
		return err
	}

	for key, value := range mapToGitCommitHash {
		fmt.Println(key)
		fmt.Println(value)
	}

	/*data, err := db.Get([]byte(gitCommitHash), nil)

	ok, err := db.Has([]byte(gitCommitHash), nil)
	if err != nil {
		fmt.Println(err)
	}
	if ok {
		fmt.Println("present-----")
	}
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("here is data", data)*/

	return nil
}

func OpenGolevelDB() (*leveldb.DB, error) {
	path := filepath.Join(os.TempDir(), AppName)

	if _, err := os.Stat(path); err != nil {
		err := os.Mkdir(path, 0755)
		if err != nil {
			return nil, err
		}
	}
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}
