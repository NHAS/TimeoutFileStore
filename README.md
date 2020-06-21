# TimeoutFileStore
Serve a file for 1 hour, behind authentication. 

Worth noting that while this follows good practice for file uploading (content-disposition header, guids for resources) it lacks some security controls. 

* Lack of maximum file upload size
* Lack of maximum content stored per user
* GORM may or may not support sqlite3 with multiple threads, but its hard to tell (may cause race conditions under very specific circumstances) 
* Lack of CSRF tokens (cookies are set as samesite) 

It is also missing three features which would be rather helpful.

* Ability to set file expiry
* Ability to set file public/private
* Admin users being able to access their own files (more of a UI bug than anything due to how I've set up templating)
