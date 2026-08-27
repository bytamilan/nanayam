# API கையேடு

**மொழிகள்:** [English](API-Reference) · **தமிழ்**

Gateway ஒரே செயல்பாடுகளை REST (`:8080`) மற்றும் gRPC (`:50051`) இரண்டின் மூலமும் வெளிப்படுத்துகிறது.

---

## அங்கீகாரம்

`/health`, `/v1/Config`, `/v1/Login`, `/v1/Register` தவிர ஒவ்வொரு endpoint-க்கும் bearer டோக்கன் தேவை.

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/v1/Login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin"}' | jq -r .token)

curl http://localhost:8080/v1/ListAssets -H "Authorization: Bearer $TOKEN"
```

டோக்கன்கள் `sub`, `usr`, `org`, `role`, `iat`, `exp` ஆகியவற்றைச் சுமக்கும் HS256 JWT-க்கள். `AUTH_SESSION_HOURS` (இயல்பாக 24) நேரத்திற்குப் பிறகு காலாவதியாகும்.

இல்லாத, தவறான வடிவிலான, காலாவதியான, அல்லது வேறு சாவியால் கையொப்பமிடப்பட்ட டோக்கன்கள் அனைத்தும் **401** திருப்பித் தரும்.

---

## பொது endpoint-கள்

### `GET /health`

```json
{ "status": "ok" }
```

Kubernetes readiness மற்றும் liveness probes இதைப் பயன்படுத்துகின்றன.

### `GET /v1/Config`

```json
{ "signupEnabled": false, "channel": "mychannel", "chaincode": "basic" }
```

பதிவு இணைப்பைக் காட்ட வேண்டுமா என்று முடிவெடுக்க கன்சோல் இதைப் படிக்கிறது. Gateway-ஐ அடைய முடியாவிட்டால், கன்சோல் `signupEnabled`-ஐ `false` எனக் கருதுகிறது — பாதுகாப்பான பக்கம் நோக்கித் தோல்வியடைகிறது.

### `POST /v1/Login`

```json
{ "username": "admin", "password": "admin" }
```

**200** `{ "token": "eyJhbGci…" }` · தவறான சான்றுகளுக்கு **401**.

பயனர்பெயர் தெரியாததாக இருந்தாலும், கடவுச்சொல் தவறாக இருந்தாலும் பிழைச் செய்தி ஒன்றேதான். எனவே இந்த endpoint-ஐ வைத்துக் கணக்குகளைக் கண்டறிய முடியாது.

### `POST /v1/Register`

```json
{ "username": "alice", "password": "s3cret", "org": "Org1MSP" }
```

உருவாக்கப்பட்ட பயனருடன் **201** · `AUTH_SIGNUP_ENABLED` `true` இல்லாதபோது **403** · நகல் அல்லது காலியான பயனர்பெயருக்கு **400**.

கடவுச்சொல் hash பதிலில் ஒருபோதும் சேர்க்கப்படுவதில்லை.

---

## அங்கீகாரம் தேவைப்படும் endpoint-கள்

### `GET /v1/Me`

அங்கீகரிக்கப்பட்ட பயனரைத் திருப்பித் தருகிறது: `id`, `username`, `org`, `role`, `createdAt`.

### `GET /v1/ChannelInfo`

சேனல் பெயர், chaincode பெயர், MSP id, சேனலில் உள்ள நிறுவனங்களும் அவற்றின் பங்குகளும், orderer endpoint-கள்.

### Assets

| Endpoint | முறை | Body / Query | திருப்பித் தருவது |
|---|---|---|---|
| `/v1/CreateAsset` | POST | `assetId`, `color`, `size`, `owner`, `appraisedValue` | `{ "success": true }` |
| `/v1/QueryAsset` | GET | `?assetId=` | `{ "data": "…json…" }` |
| `/v1/ListAssets` | GET | — | `{ "assetIds": [...] }` |

```bash
curl -X POST http://localhost:8080/v1/CreateAsset \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"assetId":"asset1","color":"blue","size":5,"owner":"Alice","appraisedValue":300}'
```

`CreateAsset` ஒரு பரிவர்த்தனையைச் சமர்ப்பித்து அது commit ஆகும் வரை காத்திருக்கிறது; எனவே இது ஏறக்குறைய ஒரு பிளாக் நேரம் எடுக்கும். Query endpoint-கள் ஒரே பியரை மட்டும் மதிப்பிட்டு மில்லிவினாடிகளில் திரும்பும்.

### புகார்கள்

| Endpoint | முறை | Body / Query | திருப்பித் தருவது |
|---|---|---|---|
| `/v1/SubmitComplaint` | POST | `complaintId`, `category`, `citizenHash`, `descriptionHash`, `attachmentsRef` | `{ "success": true }` |
| `/v1/UpdateComplaint` | POST | `complaintId`, `action`, `value` | `{ "success": true }` |
| `/v1/QueryComplaint` | GET | `?complaintId=` | `{ "data": "…json…" }` |
| `/v1/ListComplaints` | GET | — | `{ "complaintIds": [...] }` |
| `/v1/GetComplaintHistory` | GET | `?complaintId=` | `{ "data": "…json array…" }` |

பேரேட்டில் ஹாஷ்கள் மட்டுமே செல்கின்றன — `citizenHash`, `descriptionHash`; பெயர்களோ புகார் உரையோ அல்ல. தனிநபரை அடையாளம் காட்டும் தரவு சங்கிலிக்கு வெளியே இருக்கிறது; பதிவு மாறவில்லை என்பதைப் பேரேடு நிரூபிக்கிறது — அதன் உள்ளடக்கத்தைச் சேனலில் உள்ள ஒவ்வொரு நிறுவனத்திற்கும் வெளியிடாமல்.

`GetComplaintHistory` பதிவின் ஒவ்வொரு பதிப்பையும், அதன் பரிவர்த்தனை id மற்றும் நேர முத்திரையுடன் திருப்பித் தருகிறது. அந்த வரலாற்றைத் திருத்த முடியாது — புகார் நடைமுறையைப் பேரேட்டில் வைப்பதன் முழு நோக்கமும் இதுவே.

### பேரேடு உலாவி

| Endpoint | முறை | Query | திருப்பித் தருவது |
|---|---|---|---|
| `/v1/LedgerBlocks` | GET | `?start=0&end=10` | பிளாக் சுருக்கங்கள் |
| `/v1/LedgerActivity` | GET | — | சங்கிலி உயரமும் எண்ணிக்கைகளும் |

---

## நிலைக் குறியீடுகள்

| குறியீடு | பொருள் |
|---|---|
| 200 | வெற்றி |
| 201 | பயனர் உருவாக்கப்பட்டார் |
| 400 | தவறான JSON, அல்லது தேவையான அளவுரு இல்லை |
| 401 | இல்லாத, தவறான, காலாவதியான, அல்லது செல்லாத டோக்கன் |
| 403 | பதிவு முடக்கப்பட்ட நிலையில் பதிவு முயற்சி |
| 405 | அந்த endpoint-க்குத் தவறான HTTP முறை |
| 500 | Chaincode அல்லது பியர் பிழையைத் திருப்பித் தந்தது |
| 502 | கன்சோலால் பகுக்க முடியாத ஒன்றை gateway திருப்பித் தந்தது |

---

## gRPC

சேவை `proto/fabric.proto`-இல் வரையறுக்கப்பட்டுள்ளது:

```protobuf
service FabricService {
  rpc CreateAsset(CreateAssetRequest) returns (CreateAssetResponse);
  rpc QueryAsset(QueryAssetRequest) returns (QueryAssetResponse);
  rpc ListAssets(ListAssetsRequest) returns (ListAssetsResponse);
  rpc SubmitComplaint(SubmitComplaintRequest) returns (SubmitComplaintResponse);
  rpc UpdateComplaint(UpdateComplaintRequest) returns (UpdateComplaintResponse);
  rpc QueryComplaint(QueryComplaintRequest) returns (QueryComplaintResponse);
  rpc ListComplaints(ListComplaintsRequest) returns (ListComplaintsResponse);
  rpc GetComplaintHistory(GetComplaintHistoryRequest) returns (GetComplaintHistoryResponse);
}
```

```bash
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext -d '{}' localhost:50051 FabricService/ListAssets
```

---

## CORS

Gateway எல்லா origin-களையும் அனுமதிக்கிறது; இது மேம்பாட்டுக்குப் பொருத்தமானது. வேறு யாராவது அணுகக்கூடிய நிறுவலுக்கு, உண்மையான origin கொள்கையை அமைக்கும் ingress-இல் முடிக்கவும், அல்லது `services/gateway/http.go`-இல் உள்ள `corsMiddleware`-ஐக் குறுக்கவும்.
