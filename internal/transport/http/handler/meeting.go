package handler
import "net/http"
// implement me
type MeetingHandler struct{}
func NewMeetingHandler() *MeetingHandler {
// implement me
return &MeetingHandler{}
}
func (h *MeetingHandler) List(w http.ResponseWriter, r *http.Request) {
// implement me
}
func (h *MeetingHandler) Create(w http.ResponseWriter, r *http.Request) {
// implement me
}
func (h *MeetingHandler) GetByID(w http.ResponseWriter, r *http.Request) {
// implement me
}
func (h *MeetingHandler) ExportAgenda(w http.ResponseWriter, r *http.Request) {
// implement me
}
func (h *MeetingHandler) ExportParticipants(w http.ResponseWriter, r *http.Request) {
// implement me
}
