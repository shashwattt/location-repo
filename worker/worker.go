package main

import (
    "net/http"
    "fmt"
    "encoding/json"
    "log"
    "golang.org/x/net/context"
    "cloud.google.com/go/pubsub"
    "time"
    )

type RerquiredInfo struct{
    Country         string
    Country_code    string
    Region          string
    Region_code     string
    Zip_code        string         
}

func main(){

	ctx := context.Background()
    client, err := pubsub.NewClient(ctx, "pretlist-daemons-apps-us-east1")
    if err != nil {
        log.Fatalf("Could not create pubsub Client: %v", err)
    }

    const topicname = "loc-service"
    topic := createTopicIfNotExists(client, topicname) 
	
	const subName = "loc-subs-test"
	
	if err := create(client, subName, topic); err != nil {
		log.Fatal(err)
	}
    if err := pullMsgs(client, subName, topic); 	
    err != nil {
		log.Fatal(err)
	}
	// Delete the subscription.
	if err := delete(client, subName); err != nil {
		log.Fatal(err)
	}
	err = http.ListenAndServe(":8081", nil)
    if err != nil {
        log.Println("ListenAndServe: "+err.Error())
    }else{
    	 log.Println("Listening on: 8081")
    }
}


func createTopicIfNotExists(c *pubsub.Client, topicName string) *pubsub.Topic {
	ctx := context.Background()

	t := c.Topic(topicName)
	ok, err := t.Exists(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if ok {
		return t
	}

	t, err = c.CreateTopic(ctx, topicName)
	if err != nil {
		log.Fatalf("Failed to create the topic: %v", err)
	}
	return t
}

func pullMsgs(client *pubsub.Client, name string, topic *pubsub.Topic) error {
	ctx := context.Background()
	sub := client.Subscription(name)
	it, err := sub.Pull(ctx)
	if err != nil {
		return err
	}
	defer it.Stop()
	
	for {
		log.Println("In for loop")
		msg, err := it.Next()
		if err != nil {
			log.Fatalf("could not pull: %v", err)
		}
		var reqInfo RerquiredInfo
		if err := json.Unmarshal(msg.Data, &reqInfo); err != nil {
			log.Printf("could not decode message data: %#v", msg)
			msg.Done(true)
			continue
		}

		//log.Printf("[publishedData %d] Processing.", reqInfo.Country_code)
		log.Println("You are in- " + reqInfo.Region +", "+ reqInfo.Region_code +", "+ reqInfo.Country+", "+ reqInfo.Country_code +", "+ reqInfo.Zip_code);
	}
	return nil
}

func create(client *pubsub.Client, name string, topic *pubsub.Topic) error {
	ctx := context.Background()
	// [START create_subscription]
	sub, err := client.CreateSubscription(ctx, name, topic, 20*time.Second, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Created subscription: %v\n", sub)
	// [END create_subscription]
	return nil
}

func delete(client *pubsub.Client, name string) error {
	ctx := context.Background()
	// [START delete_subscription]
	sub := client.Subscription(name)
	if err := sub.Delete(ctx); err != nil {
		return err
	}
	fmt.Println("Subscription deleted.")
	
	return nil
}