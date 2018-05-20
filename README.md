# amz-scrapper-go

## build & run

`docker build -t app-img .`

for development image: `docker build --target=build -t app-img-build .` 

`docker run --rm -it -p 8080:8080 app-img`

You should be able to access service at localhost:8080 or `$(docker machine ip)`:8080

## usage 

```bash
curl -i -X POST -H "Content-Type: application/json" http://localhost:8080  \
-d '["https://www.amazon.co.uk/gp/product/1509836071","https://www.amazon.co.uk/gp/product/1509836072"]' 
```

```
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sun, 20 May 2018 20:59:11 GMT
Content-Length: 349

[{"url":"https://www.amazon.co.uk/gp/product/1509836071","meta":{"title":"The Fat-Loss Plan: 100 Quick and Easy Recipes with Workouts","price":"£8.49","image":"https://images-na.ssl-images-amazon.com/images/I/51IsTylYiPL._SX382_BO1,204,203,200_.jpg","in_stock":true}},{"url":"https://www.amazon.co.uk/gp/product/1509836072wqe","error":"Not Found"}]
```

### async

You can POST to /:requestID so you can later access it with GET /:requestID

```bash
curl -i -X POST -H "Content-Type: application/json" http://localhost:8080/request-id  \
-d '["https://www.amazon.co.uk/gp/product/1509836071","https://www.amazon.co.uk/gp/product/1509836072"]' 
```

```
HTTP/1.1 201 Created
Date: Sun, 20 May 2018 21:02:02 GMT
Content-Length: 0
```

GET /:requestID will block until job is done, call to /:requestID will also pop job out of storage, so subsequent calls will return 404

```bash
curl -i -X GET http://localhost:8080/request-id
```

```
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sun, 20 May 2018 21:02:51 GMT
Content-Length: 349

[{"url":"https://www.amazon.co.uk/gp/product/1509836071","meta":{"title":"The Fat-Loss Plan: 100 Quick and Easy Recipes with Workouts","price":"£8.49","image":"https://images-na.ssl-images-amazon.com/images/I/51IsTylYiPL._SX382_BO1,204,203,200_.jpg","in_stock":true}},{"url":"https://www.amazon.co.uk/gp/product/1509836072wqe","error":"Not Found"}]
```

```bash
curl -i -X GET http://localhost:8080/request-id
```

```
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Sun, 20 May 2018 21:58:24 GMT
Content-Length: 19

404 page not found
```






