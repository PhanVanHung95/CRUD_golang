Project use restful api server only with **gorilla/mux**.  

## Install and Run
```shell
## RUN
- go run server.go
- go run rabbitmq_worker.go
## API Endpoint
- http://localhost:3000/api/v1/companies
    - `GET`: get list of companies
    - `POST`: create company
- http://localhost:3000/api/v1/companies/{name}
    - `GET`: get company
    - `PUT`: update company
    - `DELETE`: remove company

## Data Structure
```json
{
  "name": "VIEON",
  "tel": "012-345-6789",
  "email": "VIEON@datvietVAC.com"
}
```