# Vespa Application

벡터 검색을 위한 Vespa 로컬 개발 환경.

---

## 폴더 구조

```
vespa-app/
├── docker-compose.yml   # 컨테이너 설정
├── services.xml         # Vespa 서비스 설정
├── schemas/
│   └── sample_vector.sd # 벡터 스키마 정의
└── README.md
```

---

## 기동 방법

### 1. 컨테이너 시작

```bash
docker-compose up -d
```

### 2. Config Server 준비 확인

```bash
curl http://localhost:19071/state/v1/health
# {"status":{"code":"up"}} 확인
```

### 3. Application 배포

```bash
# 방법 1: curl (vespa CLI 없이)
zip -r - schemas services.xml | curl --header "Content-Type: application/zip" \
  --data-binary @- http://localhost:19071/application/v2/tenant/default/prepareandactivate

# 방법 2: vespa CLI
brew install vespa-cli
vespa config set target local
vespa deploy --wait 300 .
```

### 4. 서비스 확인

```bash
curl http://localhost:8080/state/v1/health
# {"status":{"code":"up"}} 확인
```

---

## 포트

| 포트 | 용도 | 설명 |
|------|------|------|
| **19071** | Config Server | Application 배포, 설정 관리 |
| **8080** | Query/Feed API | 검색, 문서 CRUD |

### 왜 포트가 2개인가?

Vespa는 **관리 영역**과 **서비스 영역**을 분리한다.

```
┌─────────────────────────────────────────────────────────┐
│  19071 (Config Server)                                  │
│  - 스키마/설정 배포 (vespa deploy)                        │
│  - 클러스터 관리                                         │
│  - 배포 후에는 거의 사용 안 함                             │
│  - 프로덕션에서는 외부 노출 X (관리자만 접근)               │
└─────────────────────────────────────────────────────────┘
                         │
                         │ 배포하면
                         ▼
┌─────────────────────────────────────────────────────────┐
│  8080 (Container)                                       │
│  - 문서 저장/수정/삭제 (Feed)                             │
│  - 검색 (Query)                                         │
│  - 애플리케이션이 실제로 사용하는 포트                      │
│  - 프로덕션에서 외부 노출 O                               │
└─────────────────────────────────────────────────────────┘
```

---

## Feed vs Query

| 구분 | Feed | Query |
|------|------|-------|
| **목적** | 데이터 저장/수정/삭제 | 데이터 검색 |
| **API** | `/document/v1/...` | `/search/` |
| **비유** | MySQL INSERT/UPDATE/DELETE | MySQL SELECT |
| **방향** | 앱 → Vespa (쓰기) | 앱 → Vespa → 앱 (읽기) |

### Feed (문서 입력)

```
앱 ──[문서 데이터]──▶ Vespa 저장소
```

- 벡터 + 메타데이터를 Vespa에 저장
- 실시간 또는 배치로 대량 입력 가능
- 저장 즉시 검색 가능 (near real-time)

### Query (검색)

```
앱 ──[검색 조건]──▶ Vespa ──[결과]──▶ 앱
```

- 저장된 문서에서 조건에 맞는 것 찾기
- 벡터 유사도 검색, 텍스트 검색, 필터링 등

---

## API 사용법

### 문서 저장 (Feed)

```bash
curl -X POST "http://localhost:8080/document/v1/sample_vector/sample_vector/docid/doc1" \
  -H "Content-Type: application/json" \
  -d '{
    "fields": {
      "id": "doc1",
      "text": "hello world",
      "embedding": [0.1, 0.2, ..., 0.9]  # 1024차원 벡터
    }
  }'
```

### 문서 조회 (Get)

```bash
curl "http://localhost:8080/document/v1/sample_vector/sample_vector/docid/doc1"
```

### 문서 삭제 (Delete)

```bash
curl -X DELETE "http://localhost:8080/document/v1/sample_vector/sample_vector/docid/doc1"
```

### 벡터 검색 (Query)

```bash
curl -X POST "http://localhost:8080/search/" \
  -H "Content-Type: application/json" \
  -d '{
    "yql": "select * from sample_vector where {targetHits:10}nearestNeighbor(embedding, q)",
    "ranking.profile": "default",
    "input.query(q)": [0.1, 0.2, ..., 0.9],
    "hits": 10
  }'
```

### 텍스트 검색 (BM25)

```bash
curl -X POST "http://localhost:8080/search/" \
  -H "Content-Type: application/json" \
  -d '{
    "yql": "select * from sample_vector where text contains \"hello\"",
    "hits": 10
  }'
```

### 하이브리드 검색 (벡터 + 텍스트)

```bash
curl -X POST "http://localhost:8080/search/" \
  -H "Content-Type: application/json" \
  -d '{
    "yql": "select * from sample_vector where {targetHits:10}nearestNeighbor(embedding, q) or text contains \"hello\"",
    "ranking.profile": "default",
    "input.query(q)": [0.1, 0.2, ..., 0.9],
    "hits": 10
  }'
```

---

## Document API URL 구조

```
/document/v1/{namespace}/{document-type}/docid/{document-id}
```

- **namespace**: content id (services.xml의 content id)
- **document-type**: 스키마 이름 (sample_vector)
- **document-id**: 문서 고유 ID

---

## 주요 명령어

```bash
# 컨테이너 시작
docker-compose up -d

# 컨테이너 중지
docker-compose down

# 로그 확인
docker logs vespa -f

# 컨테이너 재시작 (설정 변경 후)
docker-compose down && docker-compose up -d
```

---

## 참고

- [Vespa Documentation](https://docs.vespa.ai/)
- [Vespa Query Language (YQL)](https://docs.vespa.ai/en/query-language.html)
- [Document API](https://docs.vespa.ai/en/document-v1-api-guide.html)