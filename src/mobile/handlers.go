package mobile

/*
These types are exported and need to be implemented and used by the mobile
application.
*/

// CommitHandler is called when Lachesis has committed a block to the DAG and publishes
// that message to the mobile app. It returns the state hash resulting from applying 
// the block's transactions to the state.
type CommitHandler interface {
	OnCommit([]byte) []byte
}

// Handles mobile app mobile app exceptions.
type ExceptionHandler interface {
	OnException(string)
}
