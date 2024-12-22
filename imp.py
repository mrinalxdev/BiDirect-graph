import requests
import json

# Sample connections data
connections = [
    {"sourceId": 1, "destinationIds": [2, 3, 4]},
    {"sourceId": 2, "destinationIds": [1, 3, 5]},
    {"sourceId": 3, "destinationIds": [1, 2, 4, 6]},
    {"sourceId": 4, "destinationIds": [1, 3, 7]},
    {"sourceId": 5, "destinationIds": [2, 8]},
    {"sourceId": 6, "destinationIds": [3, 9]},
    {"sourceId": 7, "destinationIds": [4, 10]}
]

for conn in connections:
    response = requests.post(
        "http://localhost:8080/api/connections",
        json=conn
    )
    print(f"Stored connections for member {conn['sourceId']}")