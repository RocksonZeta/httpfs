package async

import (
	"encoding/binary"
	"encoding/json"
	"httpfs/base"
	"httpfs/base/log"

	"github.com/boltdb/bolt"
)

type Task struct {
	Id     int
	State  int //process state
	Module string
	Method string
	Args   string
}

const (
	stateNotStart = 0
	stateStarted  = 1
	stateFinished = 2
	stateFailed   = 3
)

type TaskReslt struct {
	Result string
	TaskId int
	State  int
}

type TaskActor struct {
	TaskChan   chan Task
	nextChan   chan bool
	FinishChan chan TaskReslt
	db         *bolt.DB
	taskCount  int
}

func NewTaskActor() *TaskActor {
	r := &TaskActor{TaskChan: make(chan Task), nextChan: make(chan bool), FinishChan: make(chan TaskReslt)}
	var err error
	r.db, err = bolt.Open(base.Config.Fs.Tasks, 0644, nil)
	if err != nil {
		panic(err)
	}
	r.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("task_new"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("task_finished"))
		return err
	})
	go r.loop()
	return r
}

func (t *TaskActor) loop() {
	for {
		select {
		case task := <-t.TaskChan:
			t.save(task)
			t.sendNext()
		case <-t.nextChan:
			t.next()
		case result := <-t.FinishChan:
			t.taskCount--
			t.onFinish(result.TaskId, result.State)
			t.sendNext()
		}
	}

}
func (t *TaskActor) sendNext() {
	log.Log.Debug("TaskActor.sendNext")
	go func() {
		t.nextChan <- true
	}()
}
func (t *TaskActor) next() error {
	log.Log.Debug("TaskActor.next")
	if t.taskCount >= 1 {
		return nil
	}
	return t.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("task_new"))
		c := b.Cursor()
		k, v := c.First()
		var task Task
		if k == nil {
			return nil
		}
		err := json.Unmarshal(v, &task)
		if err != nil {
			return err
		}
		if h, ok := handlers[task.Module]; ok {
			t.taskCount++
			err := h.Do(task.Method, task.Args, task.Id, t.FinishChan)
			if err != nil {
				log.Log.Error("exec task error.", err)
			}
		}
		log.Log.Debug("handlers:", handlers, handlers[task.Module], task)
		return nil
	})
	// return nil
}

func (t *TaskActor) save(task Task) error {
	log.Log.Debug("TaskActor.save - ", task)
	return t.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("task_new"))
		id, err := b.NextSequence()
		if err != nil {
			log.Log.Error("TaskActor.save - NextSequence,", id)
			return err
		}
		task.Id = int(id)
		buf, err := json.Marshal(task)
		if err != nil {
			return err
		}
		return b.Put(itob(task.Id), buf)
	})
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// func (t *TaskActor) execute(task Task) {
// 	h, _ := handle.Get(task.Module)
// 	h.Do(task.Method, task.Args)
// }

// func (t *TaskActor) update() {

// }

func (t *TaskActor) onFinish(taskId, state int) error {
	log.Log.Debug("TaskActor.onFinish - ", taskId)
	return t.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("task_new"))
		bs := b.Get(itob(taskId))
		var task Task
		err := json.Unmarshal(bs, &task)
		if err != nil {
			return err
		}
		task.State = state
		f := tx.Bucket([]byte("task_finished"))
		buf, err := json.Marshal(task)
		if err != nil {
			return err
		}
		f.Put(itob(task.Id), buf)
		return b.Delete(itob(task.Id))
	})
}

// func (t *TaskActor) onProgress() {

// }
