package label

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	defaultLabels = "default"
)

//FindLabels find label doc by labels and project, check if the project has certain labels
//if map is empty. will return default labels doc which has no labels
func FindLabels(ctx context.Context, domain, project string, labels map[string]string) (*model.LabelDoc, error) {
	c, err := session.GetClient()
	if err != nil {
		return nil, err
	}
	collection := c.Database(session.Name).Collection(session.CollectionLabel)

	filter := bson.M{"domain": domain, "project": project}
	for k, v := range labels {
		filter["labels."+k] = v
	}
	if len(labels) == 0 {
		filter["labels"] = defaultLabels //allow key without labels
	}
	cur, err := collection.Find(ctx, filter)
	if err != nil {

		return nil, err
	}
	defer cur.Close(ctx)
	if cur.Err() != nil {
		return nil, err
	}
	openlogging.Debug(fmt.Sprintf("find labels [%s] in [%s]", labels, domain))
	curLabel := &model.LabelDoc{} //reuse this pointer to reduce GC, only clear label
	//check label length to get the exact match
	for cur.Next(ctx) { //although complexity is O(n), but there won't be so much labels
		curLabel.Labels = nil
		err := cur.Decode(curLabel)
		if err != nil {
			openlogging.Error("decode error: " + err.Error())
			return nil, err
		}
		if len(curLabel.Labels) == len(labels) {
			openlogging.Debug("hit exact labels")
			curLabel.Labels = nil //exact match don't need to return labels
			return curLabel, nil
		}

	}
	return nil, session.ErrLabelNotExists
}

//GetLatestLabel query revision table and find maximum revision number
func GetLatestLabel(ctx context.Context, labelID string) (*model.LabelRevisionDoc, error) {
	c, err := session.GetClient()
	if err != nil {
		return nil, err
	}
	collection := c.Database(session.Name).Collection(session.CollectionLabelRevision)

	filter := bson.M{"label_id": labelID}

	cur, err := collection.Find(ctx, filter,
		options.Find().SetSort(map[string]interface{}{
			"revision": -1,
		}), options.Find().SetLimit(1))
	if err != nil {
		return nil, err
	}
	h := &model.LabelRevisionDoc{}
	var exist bool
	for cur.Next(ctx) {
		if err := cur.Decode(h); err != nil {
			openlogging.Error("decode to KVs error: " + err.Error())
			return nil, err
		}
		exist = true
		break
	}
	if !exist {
		return nil, service.ErrRevisionNotExist
	}
	return h, nil
}

//Exist check whether the project has certain label or not and return label ID
func Exist(ctx context.Context, domain string, project string, labels map[string]string) (primitive.ObjectID, error) {
	l, err := FindLabels(ctx, domain, project, labels)
	if err != nil {
		if err.Error() == context.DeadlineExceeded.Error() {
			openlogging.Error("find label failed, dead line exceeded", openlogging.WithTags(openlogging.Tags{
				"timeout": session.Timeout,
			}))
			return primitive.NilObjectID, fmt.Errorf("operation timout %s", session.Timeout)
		}
		return primitive.NilObjectID, err
	}

	return l.ID, nil

}

//CreateLabel create a new label
func CreateLabel(ctx context.Context, domain string, labels map[string]string, project string) (*model.LabelDoc, error) {
	c, err := session.GetClient()
	if err != nil {
		return nil, err
	}
	l := &model.LabelDoc{
		Domain:  domain,
		Labels:  labels,
		Project: project,
	}
	collection := c.Database(session.Name).Collection(session.CollectionLabel)
	res, err := collection.InsertOne(ctx, l)
	if err != nil {
		return nil, err
	}
	objectID, _ := res.InsertedID.(primitive.ObjectID)
	l.ID = objectID
	return l, nil
}
