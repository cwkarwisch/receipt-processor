# receipt-processor

If you want to run the application in docker, you can clone the repo and from the working directory run `docker build --tag receipt-processor .`

Once the image is built, you can run the container with `docker run --name receipt-processor -p 8080:8080 receipt-processor`. The application will be listening on port 8080 of the docker container, which is exposed in the Dockerfile, but feel free to map a different port of your local to the exposed port on the container. Just use `docker run --name receipt-processor -p <your_favorite_port>:8080 receipt-processor`.

With the container running, you can then use curl to hit the two endpoints. The commands below assume you mapped your local machine's port 8080 to the container's port 8080 when starting the container.

```
curl http://localhost:8080/receipts/process -d '{"retailer": "Walgreens", "purchaseDate": "2022-01-02", "purchaseTime": "08:13", "total": "2.65", "items": [{"shortDescription": "Pepsi - 12-oz", "price": "1.25"}, {"shortDescription": "Dasani", "price": "1.40"}]}' -v -H "Content-Type: application/json"
```

Grab the uuid sent in response, and then send the following:

`curl -X GET http://localhost:8080/receipts/{uuid_you_just_grabbed}/points -v`