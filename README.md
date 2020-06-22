# TimeoutFileStore
Serve a file for 1 hour, behind authentication. 

Worth noting that while this follows good practice for file uploading (content-disposition header, guids for resources) it lacks some security controls. 

* Lack of maximum file upload size
* Lack of maximum content stored per user
* GORM may or may not support sqlite3 with multiple threads, but its hard to tell (may cause race conditions under very specific circumstances) 
* Self made authentication token mechanism, which may be vulnerable to god knows what, I've tried to make it as secure as possible. But meh.

It is also missing some features which would be rather helpful:

* Change password
* Purge all files
* Set file public/private
