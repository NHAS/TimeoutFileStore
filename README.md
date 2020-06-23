# TimeoutFileStore
Store files for a set amount of time. 

## Use case
Ever want to travel somewhere but not super sure about those border controls siezing your computers, forcing you to unlock them and using them to take control of your infastructure? 
Nope?  
Me neither. However I think its a fun thought experiment.  

This tool goes part way to solving this issue by creating a "stall" tactic. If you stall long enough, your files which in this case would be 'keys' to your infastructure will be gone and rendering you unable to give up any way of accessing your computers. 

Be warned. I have no idea how this would go legally for you. Nor how it would go physically. 

But again, thought experiment. 

## Installation

```
go get -u https://github.com/NHAS/TimeoutFileStore
cd $GOHOME/go/src/github.com/NHAS/TimeoutFileStore/
go build .
```

## Screenshots

**Login**
![Login](/images/login.png?raw=true)
</br>

**User List**
![User List](/images/user_list.png?raw=true)
</br>

**File List**
![File List](/images/files_list.png?raw=true)
</br>
## Things to keep in mind

While this follows good practice for file uploading (content-disposition header, guids for resources) it lacks some security controls. 

* Lack of maximum file upload size
* Lack of maximum content stored per user
* GORM may or may not support sqlite3 with multiple threads, but its hard to tell (may cause race conditions under very specific circumstances) 
* Self made authentication token mechanism, which may be vulnerable to god knows what, I've tried to make it as secure as possible. But meh.

It is also missing some features which would be rather helpful:

* Change password
* Purge all files
* Set file public/private
