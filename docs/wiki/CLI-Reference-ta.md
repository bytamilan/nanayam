---
layout: default
title: CLI கையேடு
lang: ta
---

# CLI கையேடு

**மொழிகள்:** [English](CLI-Reference.html) · **தமிழ்**

ஒவ்வொரு கட்டளையும் `--help` ஆதரிக்கிறது. பொது flag-கள்: மாற்று அமைப்பைச் சுட்ட `--config <கோப்பு>`, விரிவான வெளியீட்டுக்கு `--verbose`.

---

## `nanayam prerequisites`

நாணயத்திற்குத் தேவையான அனைத்தையும் சரிபார்த்து, விரும்பினால் நிறுவுகிறது.

```bash
nanayam prerequisites          # என்ன இல்லை என்று தெரிவிக்கிறது
nanayam prerequisites --auto   # இல்லாதவற்றை நிறுவுகிறது
```

Docker Compose, `jq`, மற்றும் Fabric பைனரிகளை `~/.nanayam/fabric-bin`-இல் நிறுவுகிறது.

---

## `nanayam network`

| கட்டளை | விளைவு |
|---|---|
| `network up` | இல்லாத artifacts-ஐ முதலில் உருவாக்கி, நெட்வொர்க்கைத் தொடங்கு |
| `network down` | கன்டெய்னர்களை நிறுத்து, volumes-ஐயும் தரவையும் வைத்திரு |
| `network clean` | கன்டெய்னர்களை நிறுத்தி எல்லாத் தரவையும் நீக்கு |
| `network status` | என்ன இயங்குகிறது என்று காட்டு |

```bash
nanayam network up                                  # அடிப்படை profile
nanayam network up --profile complaint              # புகார் நெட்வொர்க்
nanayam network up --config docker/fabric-network.yaml
```

**Profile-க்கும் `--config`-க்கும் வேறுபாடு.** Profile என்பது நாணயம் சரிசெய்யக்கூடிய அறியப்பட்ட நெட்வொர்க்: crypto அல்லது சேனல் artifacts இல்லையென்றால் அவற்றை உருவாக்கும். `--config` ஒரு குறிப்பிட்ட compose கோப்பைச் சுட்டி அதை மட்டுமே பயன்படுத்துகிறது. நாணயம் mount-களைச் சரிபார்க்கும், ஆனால் தனக்குத் தெரியாத அமைப்புக்கு crypto-வைக் கற்பனை செய்து உருவாக்காது; `apps.yaml`-ஐயும் தானாக இணைக்காது.

---

## `nanayam crypto`

```bash
nanayam crypto generate
nanayam crypto generate --config config/crypto-config-complaint.yaml --output crypto-config
nanayam crypto generate --channel-artifacts=false   # சான்றிதழ்கள் மட்டும்
nanayam crypto renew --org Org1
```

`generate` முதலில் `cryptogen`-ஐயும், வேறுவிதமாகச் சொல்லப்படாவிட்டால் genesis block, சேனல் பரிவர்த்தனை, anchor peer புதுப்பிப்புகளுக்காக `configtxgen`-ஐயும் இயக்குகிறது. எந்த crypto அமைப்பு இந்த artifacts-ஐ உருவாக்கியது என்பதை `channel-artifacts/.nanayam-artifact-source`-இல் பதிவு செய்கிறது. இதனால் பிற்பாடு `network up` புகார்-நெட்வொர்க் artifacts-ஐ அடிப்படை நெட்வொர்க்குடையவற்றிலிருந்து வேறுபடுத்தி, பொருந்தாதபோது மீண்டும் உருவாக்க முடியும்.

---

## `nanayam channel`

```bash
nanayam channel create --name mychannel --profile TwoOrgsChannel
nanayam channel join --name mychannel
nanayam channel join --name mychannel --org Org2
nanayam channel list
nanayam channel update-anchor --name mychannel --org Org1MSP
```

---

## `nanayam chaincode`

```bash
nanayam chaincode package --path ./chaincode/asset-transfer-basic --name basic --version 1.0
nanayam chaincode install --package basic.tar.gz
nanayam chaincode queryinstalled
nanayam chaincode approve --name basic --channel mychannel --package-id basic_1.0:abc123
nanayam chaincode checkcommitreadiness --name basic --channel mychannel
nanayam chaincode commit --name basic --channel mychannel
nanayam chaincode invoke --channel mychannel --name basic --function InitLedger
nanayam chaincode invoke --channel mychannel --name basic --function CreateAsset --args asset1,blue,5,Alice,300
nanayam chaincode query --channel mychannel --name basic --function GetAllAssets
```

Endorsement கொள்கையில் உள்ள ஒவ்வொரு நிறுவனத்திற்கும் `approve` ஒரு முறை இயக்கப்பட வேண்டும். எந்த நிறுவனங்கள் ஒப்புதல் அளித்தன, எவை அளிக்கவில்லை என்பதை `checkcommitreadiness` காட்டும் — `commit` தோல்வியடையும்போது முதலில் பார்க்க வேண்டியது இதுவே.

---

## `nanayam node`

```bash
nanayam node init --type peer --org Org1
nanayam node init --type orderer
nanayam node init --type ca --org Org1
nanayam node start peer0.org1.nanayam.com
nanayam node stop  peer0.org1.nanayam.com
nanayam node status
```

`node init` உட்பொதிக்கப்பட்ட டெம்ப்ளேட்களிலிருந்து ஒரே நோடுக்கான compose கோப்பையும் cryptogen அமைப்பையும் உருவாக்குகிறது — முழு நெட்வொர்க்கைத் தொடங்குவதற்குப் பதிலாக, ஏற்கனவே உள்ள கூட்டமைப்பில் ஒரு நிறுவனத்தைச் சேர்க்கும்போது பயனுள்ளது.

---

## `nanayam user`

```bash
nanayam user create --id alice --secret alicepw --type client --org Org1
nanayam user enroll --id alice --secret alicepw --org Org1
nanayam user list --org Org1
```

இவை **Fabric அடையாளங்களை** நிர்வகிக்கின்றன — CA-விலிருந்து வரும் X.509 சான்றிதழ்கள். கன்சோல் கணக்குகள் தனியானவை; [கட்டமைப்பு](Architecture-ta.html) பார்க்கவும்.

---

## `nanayam consortium`

```bash
nanayam consortium connect \
  --orderer orderer.example.com:7050 \
  --tls-cert ./tls.crt \
  --org NewOrg \
  --domain neworg.example.com

nanayam consortium join-channel --name mychannel --block ./mychannel.block
```

உங்கள் சொந்த நெட்வொர்க்கை இயக்குவதற்குப் பதிலாக, ஏற்கனவே உள்ள பல-நிறுவன நெட்வொர்க்கில் இணைவதற்கு.

---

## `nanayam gateway` மற்றும் `nanayam console`

```bash
nanayam gateway                    # REST :8080, gRPC :50051
nanayam gateway --http-port 9090

nanayam console                    # Next.js dev server :3000
nanayam console --port 4000
nanayam console --docker           # கன்டெய்னராக உருவாக்கி இயக்கு
```

---

## `nanayam upgrade`

```bash
nanayam upgrade --check                                   # புதிய வெளியீடு உள்ளதா?
nanayam upgrade                                           # சமீபத்தியதை நிறுவு
nanayam upgrade --refresh                                 # தற்போதைய வெளியீட்டை மீண்டும் நிறுவு
nanayam upgrade --dev-local --refresh --source /path/to/nanayam
```

பதிப்பு ஒப்பீடு semver அறிந்தது; எனவே `upgrade` உங்களைப் பழைய tag-க்குப் பின்னோக்கி நகர்த்தாது. மூலத்திலிருந்து உருவாக்கப்பட்ட பைனரி `dev` எனத் தெரிவிக்கிறது; tag பெற்ற எந்த வெளியீட்டையும் புதியதாகக் கருதுகிறது.

---

## Make targets

| Target | விளைவு |
|---|---|
| `make build` | CLI-ஐ `build/`-இல் உருவாக்கு |
| `make install` | `~/.nanayam/bin`-இல் நிறுவு |
| `make test` | எல்லாச் சோதனைத் தொகுப்புகளையும் இயக்கு: CLI, gateway, கன்சோல் |
| `make test-cli` / `test-gateway` / `test-console` | ஒரு தொகுப்பை மட்டும் இயக்கு |
| `make lint` | இரு Go modules-க்கும் `go vet` |
| `make fmt` / `make fmt-check` | வடிவமைக்கவும், அல்லது வடிவமைக்கப்படாவிட்டால் தோல்வியடையவும் |
| `make validate` | fmt-check, lint, build, test |
| `make build-all` | எல்லா தளங்களுக்கும் cross-compile |
| `make release-assets` | வெளியீட்டு archive-களை உருவாக்கு |
| `make deploy-cloud ARGS="..."` | கிளவுட் நிறுவல் ஸ்கிரிப்டை இயக்கு |
| `make help` | ஒவ்வொரு target-ஐயும் பட்டியலிடு |
