# In Memory Cache with Recovery

This project was based on a short tutorial and aims to explore key concepts and techniques in Golang by building a simple in memory cache with a Write-Ahead-Log.

It runs a tcp server that can accept the following commands from `redis-cli`:
```
set
get 
hset
hget
```

More features will be added eventually.

See the original article [here](https://www.build-redis-from-scratch.dev/en/introduction).

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Usage](#usage)

## Overview

This project is designed to help learners understand [specific Golang concepts, e.g., concurrency, data structures, web development, etc.]. It includes various examples and exercises to practice and solidify your knowledge.

## Installation
To get started with the project, follow these steps:

1. **Clone the repository:**

    ```sh
    git clone https://github.com/samasno/little-redis.git
    ```

2. **Navigate into the project directory:**

    ```sh
    cd little-redis
    ```

3. **Ensure you have Go 1.22 or higher installed:** [Download and install Go](https://golang.org/dl/)

4. **Install project dependencies:**
Install `redis-stack-server` to have access to the `redis-cli`.

## Usage

To run the project, use the following command:

```sh
go run *.go --port 8080 --db ./backup.db
```  

`--db` will serve as the WAL and if it already exists it will read the existing values into memory on start.

In a different terminal, run the following commands:
```
redis-cli -p 8081 set test worked // OK
redis-cli -p 8081 get test // "worked"
redis-cli -p 8081 hset h set works // OK
redis-cli -p 8081 hget h set // "works
```