# சிக்கல் தீர்வு

**மொழிகள்:** [English](Troubleshooting) · **தமிழ்**

---

## முதலில் கண்டறியுங்கள்

```bash
docker ps -a                                    # என்ன இயங்குகிறது, என்ன வெளியேறியது
docker logs peer0.org1.nanayam.com --tail 100   # ஏன் வெளியேறியது
nanayam node status
curl http://localhost:8080/health
```

பெரும்பாலான Fabric தோல்விகள் இல்லாத அல்லது தவறான சான்றிதழால் ஏற்படுகின்றன. முக்கியமான பதிவு வரி பொதுவாக **முதல்** பிழை, கடைசியல்ல.

---

## `missing or incomplete Fabric artifacts`

Docker தொடங்குவதற்கு முன் நாணயம் ஒவ்வொரு bind mount-ஐயும் சரிபார்க்கிறது; எனவே இது சரிபார்ப்பு சரியாக வேலை செய்வதன் அடையாளம். செய்தி கோப்பின் பெயரைச் சொல்கிறது:

```
missing or incomplete Fabric artifacts:
  - peer0.org1.nanayam.com: MSP directory /path/msp is missing signcerts
```

**தீர்வு:**

```bash
nanayam crypto generate
nanayam network up
```

இன்னும் தொடர்ந்தால், artifacts பழையவை — முதலிலிருந்து மீண்டும் உருவாக்குங்கள்:

```bash
nanayam network clean
nanayam crypto generate
nanayam network up
```

---

## `fabric binaries not found`

CLI முதலில் `./bin`, பிறகு `~/.nanayam/fabric-bin`, பிறகு `PATH` ஆகியவற்றில் தேடுகிறது.

```bash
nanayam prerequisites --auto
ls ~/.nanayam/fabric-bin      # peer, cryptogen, configtxgen, fabric-ca-client
```

நான்கும் இருக்க வேண்டும். சில இல்லையென்றால் பதிவிறக்கம் இடையில் நின்றுவிட்டது; கோப்பகத்தை நீக்கி மீண்டும் இயக்குங்கள்.

---

## `docker compose is not available`

நாணயம் Compose v2 (`docker compose`) அல்லது தனித்த `docker-compose` — இரண்டையும் ஏற்கிறது.

```bash
docker compose version || docker-compose version
```

இரண்டும் வேலை செய்யாவிட்டால், Compose v2-ஐ நிறுவி, Docker daemon இயங்குகிறதா என்று உறுதி செய்யுங்கள்: `docker info`.

---

## கன்டெய்னர் தொடங்கி உடனே வெளியேறுகிறது

```bash
docker logs <container> --tail 50
```

| பதிவு சொல்வது | காரணம் | தீர்வு |
|---|---|---|
| `cannot find MSP` | Crypto பொருள் இல்லை அல்லது தவறாக mount ஆகியுள்ளது | `nanayam crypto generate` |
| `failed to create ledger` | முந்தைய அமைப்பிலிருந்து பழைய volumes | `nanayam network clean` |
| `TLS handshake failed` | ஒரு பக்கத்திற்கு மட்டும் சான்றிதழ்கள் மீண்டும் உருவாக்கப்பட்டன | அனைத்தையும் மீண்டும் உருவாக்கி, எல்லாவற்றையும் மறுதொடக்கம் செய்யுங்கள் |
| `bind: address already in use` | வேறொரு செயல்முறை அந்த port-ஐ வைத்திருக்கிறது | கீழே பார்க்கவும் |

---

## Port ஏற்கனவே பயன்பாட்டில்

```bash
lsof -i :7051     # அல்லது :8080, :3000, :7050
```

அந்தச் செயல்முறையை நிறுத்துங்கள், அல்லது நாணயத்தை நகர்த்துங்கள்:

```bash
nanayam gateway --http-port 9090
nanayam console --port 4000
```

---

## சேனல் உருவாக்கம் தோல்வியடைகிறது

```
Error: got unexpected status: BAD_REQUEST -- error validating channel creation transaction
```

பொதுவாக, சேனல் பரிவர்த்தனையை உருவாக்கியதைவிட வேறொரு `configtx.yaml`-இலிருந்து genesis block உருவாக்கப்பட்டிருக்கும். இரண்டையும் சேர்த்து மீண்டும் உருவாக்குங்கள்:

```bash
rm -rf channel-artifacts
nanayam crypto generate
nanayam channel create --name mychannel --profile TwoOrgsChannel
```

எந்த crypto அமைப்பு artifacts-ஐ உருவாக்கியது என்பதை நாணயம் `channel-artifacts/.nanayam-artifact-source`-இல் பதிவு செய்கிறது; எனவே பிற்பாடு `network up` பொருந்தாமையைக் கண்டறிய முடியும் — ஆனால் கையால் திருத்தப்பட்ட அமைப்பு இன்னும் ஒத்துப்போகாமல் இருக்கக்கூடும்.

---

## Chaincode commit தோல்வியடைகிறது

```bash
nanayam chaincode checkcommitreadiness --name basic --channel mychannel
```

இது எந்த நிறுவனங்கள் ஒப்புதல் அளித்தன என்று பட்டியலிடும். `commit` வெற்றியடைய, endorsement கொள்கையில் உள்ள ஒவ்வொரு நிறுவனமும் **அதே** package id உடன் `approve` இயக்க வேண்டும்.

```bash
nanayam chaincode queryinstalled    # உண்மையான package id-ஐக் காட்டுகிறது
```

பொருந்தாத package id-தான் மிகவும் பொதுவான காரணம்: id-இல் package-இன் ஹாஷ் அடங்கியுள்ளது, எனவே chaincode-ஐ மீண்டும் உருவாக்கினால் அது மாறிவிடும்.

---

## Gateway பியருடன் இணைய முடியவில்லை

```
Failed to connect to Fabric gateway: connection error
```

| சரிபார்ப்பு | கட்டளை |
|---|---|
| பியர் இயங்குகிறதா? | `docker ps \| grep peer0` |
| Endpoint சரியா? | `echo $PEER_ENDPOINT` |
| TLS சான்றிதழ் உள்ளதா? | `ls $TLS_CERT_PATH` |
| ஒரே Docker நெட்வொர்க்கா? | `docker network inspect nanayam` |

Docker உள்ளே `localhost` அல்ல, சேவைப் பெயரைப் (`peer0.org1.nanayam.com:7051`) பயன்படுத்துங்கள்.

---

## கன்சோல் "Gateway down" காட்டுகிறது

```bash
curl http://localhost:8080/health
```

இது தோல்வியடைந்தால், gateway இயங்கவில்லை அல்லது அடைய முடியவில்லை. இது வெற்றியடைந்தும் கன்சோல் பிழை காட்டினால், கன்சோலின் `GATEWAY_URL` வேறு இடத்தைச் சுட்டுகிறது — அது தொடங்கப்பட்ட சூழலைச் சரிபாருங்கள். Docker அல்லது Kubernetes-இல் அது `localhost` அல்ல, சேவைப் பெயராக இருக்க வேண்டும்.

---

## ஒவ்வொரு கோரிக்கைக்கும் 401

| காரணம் | தீர்வு |
|---|---|
| டோக்கன் இல்லை | முதலில் உள்நுழையுங்கள்; `nanayam_token` குக்கீ அமைக்கப்பட்டுள்ளதா என்று பாருங்கள் |
| டோக்கன் காலாவதி | மீண்டும் உள்நுழையுங்கள்; 24 மணி நேரம் குறைவெனில் `AUTH_SESSION_HOURS` உயர்த்துங்கள் |
| வேறு `AUTH_JWT_SECRET` உடன் gateway மறுதொடக்கம் | ஏற்கனவே உள்ள டோக்கன்கள் செல்லாது — மீண்டும் உள்நுழையுங்கள் |
| Gateway மறுதொடக்கம் ஆனதே | பயனர் சேமிப்பு நினைவகத்தில் உள்ளது, எனவே மறுதொடக்கத்தில் கணக்குகள் இழக்கப்படுகின்றன |

கடைசியானதைத் தெளிவாகச் சொல்வது நல்லது: auth store **நினைவகத்தில்** உள்ளது. Gateway-ஐ மறுதொடக்கம் செய்தால் பதிவு செய்யப்பட்ட ஒவ்வொரு பயனரும் நீங்கி, `admin` மட்டும் மீண்டும் உருவாக்கப்படும். மேம்பாட்டுக்கு இது போதும்; வேறு எதற்கும் நிலையான சேமிப்பு தேவை.

---

## பதிவு 403 திருப்பித் தருகிறது

பதிவு இயல்பாகவே முடக்கப்பட்டுள்ளது. வேண்டுமென்றே இயக்குங்கள்:

```bash
AUTH_SIGNUP_ENABLED=true nanayam gateway
```

அல்லது நிறுவல் ஸ்கிரிப்டில்: `--enable-signup`.

---

## பாதைப் பிழையுடன் சோதனைகள் தோல்வியடைகின்றன

உங்களுடையது அல்லாத `/Users/…` அல்லது `/home/…` பாதையுடன் சோதனை தோல்வியடைந்தால், அது யாரோ ஒருவரின் கணினியை நேரடியாக எழுதி வைத்திருக்கிறது. அதற்குப் பதிலாக `t.TempDir()` அல்லது `repoRoot(t)` பயன்படுத்துங்கள் — [சோதனை](Testing-ta) பார்க்கவும்.

---

## Kubernetes pod-கள் தொடங்கவில்லை

```bash
kubectl -n nanayam get pods
kubectl -n nanayam describe pod <pod>
kubectl -n nanayam logs <pod>
```

| Pod நிலை | காரணம் | தீர்வு |
|---|---|---|
| `ImagePullBackOff` | கிளஸ்டரால் image-ஐ இழுக்க முடியவில்லை | Registry குறிப்பையும் pull secrets-ஐயும் சரிபாருங்கள் |
| `CreateContainerConfigError` | குறிப்பிடப்பட்ட Secret இல்லை | நிறுவல் ஸ்கிரிப்டை மீண்டும் இயக்குங்கள் |
| `CrashLoopBackOff` | கன்டெய்னர் தொடங்கி இறக்கிறது | பதிவுகளைப் படியுங்கள்; பொதுவாக crypto பொருள் |
| `Pending` | எந்த நோடிலும் இடமில்லை | `kubectl describe node`; resource requests சரிபாருங்கள் |

---

## முழு மீட்டமைப்பு

ஏதாவது சிக்கிக்கொண்டு, சுத்தமான தொடக்கம் வேண்டுமெனில்:

```bash
nanayam network clean
rm -rf crypto-config channel-artifacts
docker system prune -f
nanayam network up
```

இது எல்லாப் பேரேட்டுத் தரவையும் அழிக்கிறது. மேம்பாட்டு நெட்வொர்க்கில் அதுதான் நோக்கம்.

---

## இன்னும் சிக்கல் தொடர்ந்தால்

<https://github.com/bytamilan/nanayam/issues> இல் ஒரு issue திறந்து, இவற்றைச் சேர்க்கவும்:

- நீங்கள் என்ன இயக்கினீர்கள், என்ன எதிர்பார்த்தீர்கள்
- முழுப் பிழை வெளியீடு
- `nanayam version`, `docker --version`, `go version`
- உங்கள் OS மற்றும் architecture
