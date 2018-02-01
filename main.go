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

	// Test Code @TODO: Hanif
	err := os.RemoveAll(filepath.Join(os.TempDir(), AppName))
	if err != nil {
		log.Fatalln(err)
	}
	// Test Code Ends
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
		writer.Header().Set("Content-Type", "application/json")

		resp := map[string]v1beta1.EventList{}
		db, err := OpenGoLevelDB()
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()

		iter := db.NewIterator(nil, nil)

		for iter.Next() {
			// Remember that the contents of the returned slice should not be modified, and
			// only valid until the next call to Next.
			key := iter.Key()
			value := iter.Value()
			tmpEventList := &v1beta1.EventList{}
			err = json.Unmarshal(value, tmpEventList)
			resp[string(key)] = *tmpEventList
			if err != nil {
				fmt.Println("Error to Unmarshal")
			}
		}


		iter.Release()
		err = iter.Error()
		if err != nil {
			log.Fatalln(err)
		}

		res, err := json.Marshal(resp)
		if err != nil {
			fmt.Println(err)
		}
		writer.Write(res)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ProcessEvents(list *v1beta1.EventList, routine int) error {
	if list == nil {
		fmt.Println("Nil")
		return fmt.Errorf("%s", "Empty event list")
	}

	db, err := OpenGoLevelDB()
	if err != nil {
		return err
	}
	defer db.Close()

	mapToGitCommitHash := map[string]*v1beta1.EventList{}
	eventList := &v1beta1.EventList{}
	var events []*v1beta1.Event
	for _, val := range list.Items {
		fmt.Println(routine)
		_, err := json.MarshalIndent(val, "", "  ")
		if err != nil {
			log.Fatalln(err)
		}
		if val.ResponseObject != nil {
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
				ret, err := db.Has([]byte(gitCommitHash), nil)
				if !ret {
					mapToGitCommitHash[gitCommitHash] = &v1beta1.EventList{}
				} else {
					dbEventList, err := db.Get([]byte(gitCommitHash), nil)
					if err != nil {
						fmt.Println("Error to get from db...")
					}

					tmpEventList := &v1beta1.EventList{}
					err = json.Unmarshal(dbEventList, tmpEventList)
					if err != nil {
						fmt.Println("Error to Unmarshal...", err)
					}
					mapToGitCommitHash[gitCommitHash] = tmpEventList
				}
				eventList = mapToGitCommitHash[gitCommitHash]
				eventList.Items = append(eventList.Items, *eventTmp)
				mapToGitCommitHash[gitCommitHash] = eventList
			}
		}
	}
	mux.Lock()
	defer mux.Unlock()
	for key, value := range mapToGitCommitHash {
		tmpEventListByte, err := json.Marshal(value)
		fmt.Println(string(tmpEventListByte))
		if err != nil {
			log.Fatalln(err)
		}
		err = db.Put([]byte(key), tmpEventListByte, nil)
		if err != nil {
			fmt.Println("Error to PUT into DB")
		}

	}

	return nil
}

func OpenGoLevelDB() (*leveldb.DB, error) {
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
