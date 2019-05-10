# Circle

This is a simple stackoverflow-like website. This project was created to learn Go language and implementation of Mongo Db

## Prerequisites
- Go [[download page]](https://golang.org/dl/). Currently used : go version go1.11.4 windows/amd64
- Mongodb [[download page]](https://www.mongodb.com/download-center/community). Currently used : db version v4.0.5

## Folder structure

```bash
|   main.go
+---assets
|
+---src
|   +---data_model
|   |
|   +---utils
|   |
|   \---website
|       +---ask
|       |
|       +---discussion
|       |
|       +---error_pages
|       |
|       +---home
|       |
|       +---layout
|       |
|       +---login
|       |
|       +---logout
|       |
|       +---profile
|       |
|       \---register
|
\---upload
    \---userdata
```

| Folder / File  | Description |
| -------------- | ------------- |
| main.go        | First code called. Mostly do routing setup.  |
| assets         | Contains styling (css, js), image, and webfont files  |
| src            | Contains codes |
| src/datamodel  | Contains datamodel as representation of collection structore in mongodb |
| src/utils      | Utility codes |
| src/website    | Contains website modules / page |
| src/website/layout | Contains master page |
| upload/userdata | Contains files uploaded by user |
