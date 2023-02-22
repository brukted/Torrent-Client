# Torrent Client Using Golang

The project involves building a command-line torrent client that can download and upload files using the BitTorrent protocol. The client is built using the Go programming language and leverages the built-in concurrency and networking features of the language to provide fast and efficient downloads and uploads.

The torrent client is built using the following key components:

- Torrent file parser: The client is able to parse torrent files and extract the necessary information such as the tracker URL, file name, file size, and pieces.

- Tracker communication: The client communicates with the tracker to obtain a list of peers currently sharing the file. The client then connect to the peers and exchange data using the BitTorrent protocol.

- Piece management: The client downloads and uploads file pieces as needed to complete the file. The client keeps track of which pieces have been downloaded and which pieces are still needed.

- Concurrency: The client uses Go's concurrency features such as goroutines and channels to provide efficient downloads and uploads.

- Error handling: The client has a robust and handle errors such as connection timeouts, network failures, and corrupt data.

- Fault tolerance: The client is able to recover from errors and continue downloading and uploading files, by retrying failed operations and utilizing disk persisted bitfield.

Overall, the project provides a functional and efficient torrent client that can handle large file downloads with ease. The use of Go's concurrency features ensures that the client can handle multiple downloads and uploads simultaneously.


## Group Members
* Bruk Tedla
* Abel Mekonen
* Eyob Zebene
* Abdulaziz Ali
* Bahailu Abera
