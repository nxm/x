# Vocabulary Memorizer

example:
```bash
curl -X POST http://localhost:1323/translate -H 'Content-Type: application/json' -d '{"input": "robust"}'
{"en":"robust","pl":"odporny","type":"word","example":"A robust immune system is important for staying healthy.","phonetic":"ˈroʊˌbʌst","use_frequency":0.8,"difficulty":"medium"}
```

## TODO:
- [ ] Import translated words / sentences to airtable database
- [ ] Add an export to anki format (maybe it should be as a separate program or with additional to run as a cronjob)
