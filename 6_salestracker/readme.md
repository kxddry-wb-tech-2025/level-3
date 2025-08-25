# Salestracker

## Quickstart

1. Clone the repository

    ```bash
    git clone https://github.com/kxddry-wb-tech-2025/level-3
    cd level-3/6_salestracker
    ```

2. Create a file with environmental variables

    ```bash
    echo 'POSTGRES_PASSWORD=<your_postgres_password>
    POSTGRES_DB=salestracker
    POSTGRES_USER=<your_user>
    ' > .env
    ```

3. Launch the application

    ```bash
    docker-compose up -d --build
    ```

4. Access the application [here](https://localhost:8080/ui) by default.


## API Specification
API for managing sales

## Version: 1.0.0

### /items

#### POST
##### Summary:

Create a new item (date must be in the past)

##### Responses

| Code | Description |
| ---- | ----------- |
| 200 | Item created successfully |
| 400 | Bad request (invalid formatting, date in the future) |
| 500 | Internal server error (database error) |

#### GET
##### Summary:

Get all items (sorted by date)

##### Responses

| Code | Description |
| ---- | ----------- |
| 200 | Items retrieved successfully |
| 404 | Not found (When there are no items) |
| 500 | Internal server error (database error) |

### /items/{id}

#### PUT
##### Summary:

Update an item (date must be in the past)

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string (uuid) |

##### Responses

| Code | Description |
| ---- | ----------- |
| 200 | Item updated successfully |
| 400 | Bad request (invalid formatting, date in the future) |
| 404 | Not found |
| 500 | Internal server error (database error) |

#### DELETE
##### Summary:

Delete an item (date must be in the past)

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string (uuid) |

##### Responses

| Code | Description |
| ---- | ----------- |
| 202 | Item deleted successfully |
| 400 | Bad request (invalid formatting, date in the future) |
| 404 | Not found |
| 500 | Internal server error (database error) |

### /analytics

#### GET
##### Summary:

Get analytics

##### Responses

| Code | Description |
| ---- | ----------- |
| 200 | Analytics retrieved successfully |
| 404 | Not found (When there are no items) |
| 500 | Internal server error (database error) |
