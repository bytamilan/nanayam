---
layout: default
title: தொடங்குதல்
lang: ta
---

# தொடங்குதல்

**மொழிகள்:** [English](Getting-Started.html) · **தமிழ்**

இந்தப் பக்கம் காலியான கணினியிலிருந்து, தரவுடன் இயங்கும் பேரேடு வரை உங்களை அழைத்துச் செல்லும். சுமார் பதினைந்து நிமிடங்கள் — அதில் பெரும்பகுதி பதிவிறக்கங்களுக்குக் காத்திருப்பது.

---

## 1. முன்நிபந்தனைகள்

| கருவி | ஏன் | சரிபார்ப்பு |
|---|---|---|
| Docker + Compose v2 | ஒவ்வொரு Fabric நோடும் கன்டெய்னராக இயங்குகிறது | `docker compose version` |
| Go 1.21+ | CLI-யையும் gateway-யையும் உருவாக்குகிறது | `go version` |
| Node.js 20+ மற்றும் pnpm | கன்சோலை உருவாக்கி இயக்குகிறது | `node --version && pnpm --version` |
| `git`, `curl`, `jq` | அமைவு ஸ்கிரிப்ட்கள் பயன்படுத்துகின்றன | `git --version` |

இவற்றில் பெரும்பாலானவற்றை நாணயமே நிறுவித் தரும்:

```bash
nanayam prerequisites --auto
```

இந்தக் கட்டளை Hyperledger Fabric பைனரிகளையும் (`peer`, `cryptogen`, `configtxgen`, `fabric-ca-client`) `~/.nanayam/fabric-bin` கோப்பகத்திற்குப் பதிவிறக்கும். அவை இல்லாமல், crypto அல்லது சேனல்களைத் தொடும் எதுவும் வேலை செய்யாது.

---

## 2. CLI-ஐ நிறுவுங்கள்

```bash
curl -fsSL https://raw.githubusercontent.com/bytamilan/nanayam/main/install.sh | bash
```

அல்லது களஞ்சியத்திலிருந்து:

```bash
git clone https://github.com/bytamilan/nanayam.git
cd nanayam
make install
export PATH="$HOME/.nanayam/bin:$PATH"
```

நிறுவப்பட்டதா என்று உறுதி செய்யுங்கள்:

```bash
nanayam version
```

---

## 3. நெட்வொர்க்கை எழுப்புங்கள்

```bash
nanayam network up
```

இந்த ஒரு கட்டளை நிறையச் செய்கிறது — என்ன செய்கிறது என்று தெரிந்துகொள்வது நல்லது:

1. **Crypto பொருள் இருக்கிறதா என்று பார்க்கிறது.** `crypto-config/` இல்லாவிட்டால் அல்லது முழுமையடையாமல் இருந்தால், `cryptogen` மூலம் அதை உருவாக்க பொருத்தமான அமைவு ஸ்கிரிப்டை இயக்குகிறது.
2. **சேனல் artifacts இருக்கிறதா என்று பார்க்கிறது.** `channel-artifacts/`-இல் genesis block இல்லையென்றால், சரியான `configtx*.yaml` மீது `configtxgen` இயக்குகிறது.
3. **Compose கோப்பைச் சரிபார்க்கிறது.** Docker தொடங்கும் முன்பே ஒவ்வொரு bind mount-உம் சரிபார்க்கப்படுகிறது. எனவே காணாமல் போன சான்றிதழ், மூன்று வினாடிகளில் வெளியேறும் கன்டெய்னராக அல்லாமல், தெளிவான செய்தியாகத் தெரிவிக்கப்படுகிறது.
4. **தொகுப்பைத் தொடங்குகிறது** — Docker Compose மூலம்.

அடிப்படை asset நெட்வொர்க்குக்குப் பதிலாக புகார் பணிப்பாய்வு வேண்டுமெனில்:

```bash
nanayam network up --profile complaint
```

என்ன எழுந்தது என்று பாருங்கள்:

```bash
docker ps
nanayam node status
```

---

## 4. சேனலை உருவாக்கி பியர்களை இணையுங்கள்

```bash
nanayam channel create --name mychannel --profile TwoOrgsChannel
nanayam channel join --name mychannel
nanayam channel update-anchor --name mychannel --org Org1MSP
```

**சேனல்** என்பது குறிப்பிட்ட நிறுவனங்களால் பகிரப்படும் தனிப்பட்ட பேரேடு. அதில் இணையாத பியர்களால் அதன் தரவை அறவே பார்க்க முடியாது. **Anchor peer** என்பது மற்ற நிறுவனங்கள் உங்கள் மற்ற பியர்களைக் கண்டறியப் பயன்படுத்தும் பியர்.

---

## 5. Chaincode-ஐ நிறுவுங்கள்

Chaincode என்பது smart contract: பேரேட்டில் எழுத அனுமதிக்கப்பட்ட ஒரே விஷயம்.

```bash
nanayam chaincode package --path ./chaincode/asset-transfer-basic --name basic
nanayam chaincode install --package basic.tar.gz
nanayam chaincode approve --name basic --channel mychannel --package-id basic_1.0:<hash>
nanayam chaincode commit --name basic --channel mychannel
```

Commit வெற்றியடைவதற்கு முன், endorsement கொள்கையில் உள்ள **ஒவ்வொரு** நிறுவனமும் approve படியை மீண்டும் செய்ய வேண்டும். அதுதான் நோக்கம்: பகிரப்பட்ட பேரேட்டில் எந்த ஒரு தரப்பும் தனியாகக் குறியீட்டை நிறுவ முடியாது.

தரவை ஏற்றி மீண்டும் படியுங்கள்:

```bash
nanayam chaincode invoke --channel mychannel --name basic --function InitLedger
nanayam chaincode query  --channel mychannel --name basic --function GetAllAssets
```

---

## 6. Gateway-யையும் கன்சோலையும் தொடங்குங்கள்

```bash
nanayam gateway    # REST :8080, gRPC :50051
nanayam console    # Next.js :3000
```

<http://localhost:3000> திறந்து **admin / admin** மூலம் உள்நுழையுங்கள்.

> வேறு யாராவது அணுகும் நிலைக்கு வருவதற்கு முன் அந்தக் கடவுச்சொல்லை மாற்றுங்கள். பதிவு (registration) இயல்பாகவே முடக்கப்பட்டுள்ளது; உண்மையிலேயே தேவைப்படும்போது மட்டும் `AUTH_SIGNUP_ENABLED=true` மூலம் இயக்குங்கள்.

---

## 7. வேலை செய்கிறதா என்று உறுதி செய்யுங்கள்

```bash
curl http://localhost:8080/health
# {"status":"ok"}

TOKEN=$(curl -s -X POST http://localhost:8080/v1/Login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin"}' | jq -r .token)

curl -s http://localhost:8080/v1/ListAssets -H "Authorization: Bearer $TOKEN" | jq
```

கடைசி அழைப்பு உங்கள் assets-ஐத் திருப்பித் தந்தால், ஒவ்வொரு அடுக்கும் வேலை செய்கிறது: கன்சோல் → gateway → பியர் → பேரேடு.

---

## நிறுத்துதல்

```bash
nanayam network down     # நிறுத்து, தரவை வைத்திரு
nanayam network clean    # நிறுத்தி அனைத்தையும் அழி
```

`clean` பேரேடு, crypto பொருள், சேனல் artifacts அனைத்தையும் நீக்குகிறது. மேம்பாட்டு நெட்வொர்க்கில் ஏதாவது சிக்கிக்கொண்டால் இதுவே சரியான தேர்வு; வேறு எதிலும், உறுதிப்படுத்திக்கொள்ளுங்கள்.

---

## அடுத்து எங்கே

- [கட்டமைப்பு](Architecture-ta.html) — ஒவ்வொரு பகுதியும் உண்மையில் என்ன செய்கிறது
- [CLI கையேடு](CLI-Reference-ta.html) — ஒவ்வொரு கட்டளையும் flag-உம்
- [கிளவுட் நிறுவல்](Cloud-Deployment-ta.html) — அதே தொகுப்பு Kubernetes-இல்
- [சிக்கல் தீர்வு](Troubleshooting-ta.html) — மேலே உள்ள படி திட்டப்படி நடக்காதபோது
