This is a mini rock paper scissors game

The authentication is done via a local private key and JWTs.
The main controlling element is the **username** as it's part of the JWTs and is used around code as naive rbac.

Running the project:

go to the root directory and execute

```bash
docker-compose up -d <- spins up a small postgresql db
```
```bash
go run .
```
You can make request to localhost:9000

1. you need to register a user via **/registration**
   example:
   ```json
   {
    "username" : "peter_griffin",		
    "password" : "random123",
    "deposit" : 500
   }
   ```

2. login via **login** with your credentials
```json
   {
    "username" : "peter_griffin",
    "password" : "random123"
   }
```
3. you`ll get a Bearer token back, use it to authenticate for the other requests
```json
   "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJicnlhbmciLCJleHAiOjE3MTk1Nzg5Njl9.qbtAmYouJAUcEHgt1hK4HAsXwfuhPgowptTkZkRPxe0"
```
   All authenticated requests will need to have
   ```json
   "Authentication" : "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJicnlhbmciLCJleHAiOjE3MTk1Nzg5Njl9.qbtAmYouJAUcEHgt1hK4HAsXwfuhPgowptTkZkRPxe0"
   ```
(updated the token with a valid one)

- Challeging players can be done via POST **/challenge** with a **model.ChallengeRequest**
    - the choice is between 1 and 3 for **rock=1**, **paper=2**, **scissors=3**
```json
{
 "opponent" : "bryan_griffin",
 "choice" : 3,
 "bet" : 10
}
 ```
- A player can view his active pendindg challenges via GET **/challenge/pending** no need to pass anything but the Bearer token, it will get the relevant data from the db
- You can query for all players wit GET **/players** this will return all of the registered player usernames
- Accepting a challenge is done via POST **/challenge/settle** with **model.ChallengeSettleRequest**
```json
{
 "challenge_id" : "19",
 "bet_choice" : 2
}
```
- Declining is done via POST **/challenge/decline** with **model.ChallengeDeclineRequest** can be used by both challenger and oponent
```json
{
 "challenge_id" : "19"  
}
```

There's a mock implementation for transactions as I did not want to deal with real transactions, every funds change is logged in.
Players can get all the transactions they've made by querying **/transactions**

When a player challenges another, his money are removed immediatelly. In the case of declining a challenge they`re reverted.
