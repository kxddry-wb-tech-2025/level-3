// package worker is the worker that handles the delayed notifications.
// it is also responsible for cancelling bookings that are expired.
// it connects to an external microservice, which is responsible for sending notifications to users and sending us info.
package worker

type Worker struct {
}
