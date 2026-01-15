# Python Embedder 기능 명세서

> Go로 재구현하기 위한 Python Embedder 기능 분석 문서
> 원본 프로젝트: `/home/hgkim/workspace/embedding`

---

## 목차

1. [프로젝트 개요](#1-프로젝트-개요)
2. [기술 스택](#2-기술-스택)
3. [핵심 기능 목록](#3-핵심-기능-목록)
4. [상세 기능 명세](#4-상세-기능-명세)
5. [API 엔드포인트](#5-api-엔드포인트)
6. [데이터 모델](#6-데이터-모델)
7. [임베딩 전략](#7-임베딩-전략)
8. [Qdrant 벡터 DB 연동](#8-qdrant-벡터-db-연동)
9. [Go 재구현 체크리스트](#9-go-재구현-체크리스트)

---

## 1. 프로젝트 개요

| 항목         | 설명                                                |
|------------|---------------------------------------------------|
| **목적**     | 술(Alcohol) 데이터를 벡터로 임베딩하여 Qdrant에 저장하고 의미론적 검색 제공 |
| **프레임워크**  | FastAPI (Python)                                  |
| **임베딩 모델** | BGE-m3-ko (dragonkue, 1024차원, 한영 다국어)             |
| **벡터 DB**  | Qdrant (Named Vector + Sparse Vector 지원)          |
| **데이터베이스** | MySQL (SQLAlchemy ORM)                            |

---

## 2. 기술 스택

### Python 의존성

```
# 웹 프레임워크
fastapi==0.123.5
uvicorn==0.38.0
pydantic

# 데이터베이스
sqlalchemy==2.0.44
pymysql==1.1.2

# ML/임베딩
transformers==4.57.3
torch==2.9.1
sentence-transformers==5.1.2
FlagEmbedding (BGE-m3-ko)

# 벡터 DB
qdrant-client==1.16.1
```

### Go 대응 라이브러리 (추천)

| Python                | Go 대응                       |
|-----------------------|-----------------------------|
| FastAPI               | Gin / Fiber / Echo          |
| SQLAlchemy            | GORM / sqlx                 |
| qdrant-client         | qdrant-go                   |
| transformers          | HTTP API 호출 (외부 임베딩 서버)     |
| sentence-transformers | HTTP API 호출 또는 ONNX Runtime |

---

## 3. 핵심 기능 목록

### 3.1 모델 관리

| 기능            | 파일                       | 함수/메서드                           | 설명                       |
|---------------|--------------------------|----------------------------------|--------------------------|
| 모델 로드         | `config/model_loader.py` | `get_model()`                    | BGE-m3-ko 모델 로드 (LRU 캐시) |
| 디바이스 감지       | `config/model_loader.py` | `get_device()`                   | GPU/CPU 자동 감지            |
| Dense 임베딩     | `config/model_loader.py` | `embed_text(text)`               | 단일 텍스트 → 1024차원 벡터       |
| Dense 배치 임베딩  | `config/model_loader.py` | `embed_texts(texts)`             | 여러 텍스트 → 벡터 배열           |
| Sparse 임베딩    | `config/model_loader.py` | `embed_with_sparse(text)`        | Dense + Sparse 동시 추출     |
| Sparse 배치 임베딩 | `config/model_loader.py` | `embed_texts_with_sparse(texts)` | 배치 Dense + Sparse        |

### 3.2 데이터베이스

| 기능     | 파일                   | 함수/메서드                   | 설명             |
|--------|----------------------|--------------------------|----------------|
| DB 연결  | `config/database.py` | `engine`, `SessionLocal` | MySQL 연결 풀     |
| 세션 주입  | `config/database.py` | `get_db()`               | FastAPI 의존성 주입 |
| 연결 테스트 | `config/database.py` | `initialize_database()`  | 시작 시 연결 확인     |

### 3.3 Qdrant 관리

| 기능       | 파일                 | 함수/메서드                | 설명             |
|----------|--------------------|-----------------------|----------------|
| 클라이언트 생성 | `config/qdrant.py` | `qdrant_client`       | Qdrant 연결      |
| 클라이언트 주입 | `config/qdrant.py` | `get_qdrant_client()` | FastAPI 의존성 주입 |
| 연결 테스트   | `config/qdrant.py` | `initialize_qdrant()` | 시작 시 연결 확인     |

### 3.4 임베딩 변환

| 기능         | 파일                              | 함수/메서드                          | 설명                                      |
|------------|---------------------------------|---------------------------------|-----------------------------------------|
| v1 변환      | `services/embedding_service.py` | `_to_embedding_input()`         | Alcohol → AlcoholEmbeddingInput         |
| v2 변환      | `services/embedding_service.py` | `to_whisky_strategy()`          | Alcohol → WhiskyEmbeddingStrategy (4벡터) |
| 숫자 파싱      | `services/embedding_service.py` | `_parse_number()`               | "35~40" → 37.5 (범위 평균)                  |
| 범위 조회      | `services/embedding_service.py` | `get_alcohol_embeddings()`      | ID 범위로 v1 변환                            |
| 범위 조회 (v2) | `services/embedding_service.py` | `get_whisky_strategies()`       | ID 범위로 v2 변환                            |
| 페이지 조회     | `services/embedding_service.py` | `get_whisky_strategies_paged()` | OFFSET/LIMIT으로 v2 변환                    |

### 3.5 Qdrant 저장/검색

| 기능       | 파일                           | 함수/메서드                              | 설명                                 |
|----------|------------------------------|-------------------------------------|------------------------------------|
| 컬렉션 생성   | `services/qdrant_service.py` | `create_collection_if_not_exists()` | whisky_v2 컬렉션 (4 Named + 1 Sparse) |
| 단건 저장    | `services/qdrant_service.py` | `upsert_whisky()`                   | 단일 PointStruct 저장                  |
| 배치 저장    | `services/qdrant_service.py` | `upsert_whisky_batch()`             | 여러 PointStruct 저장                  |
| 하이브리드 검색 | `services/qdrant_service.py` | `get_collections_by_keyword()`      | 4개 벡터 + Sparse DBSF 융합             |
| 단일 벡터 검색 | `services/qdrant_service.py` | `search_by_vector_type()`           | 특정 벡터 타입만 검색                       |

---

## 4. 상세 기능 명세

### 4.1 임베딩 모델 (model_loader.py)

#### `get_model()` - 모델 로드

```python
@lru_cache(maxsize=1)
def get_model() -> BGEM3FlagModel:
    """
    BGE-m3-ko 모델을 로드하고 캐싱
    - 모델명: dragonkue/BGE-m3-ko
    - 벡터 차원: 1024
    - 언어: 한국어/영어 다국어
    """
```

**Go 구현 방향:**

- 임베딩 모델을 직접 로드하기 어려우므로 **외부 HTTP API** 사용 권장
- Python FastAPI 임베딩 서버를 별도로 두거나
- Hugging Face Inference API, OpenAI Embeddings API 등 활용
- 또는 ONNX 모델로 변환 후 `onnxruntime-go` 사용

#### `embed_text(text)` - 단일 임베딩

```python
def embed_text(text: str) -> tuple[np.ndarray, dict]:
    """
    단일 텍스트를 임베딩

    Returns:
        - dense_vector: np.ndarray (1024,)
        - sparse: {"indices": list[int], "values": list[float]}
    """
```

#### `embed_texts_with_sparse(texts)` - 배치 임베딩

```python
def embed_texts_with_sparse(texts: list[str]) -> tuple[list, list]:
    """
    여러 텍스트를 동시에 임베딩

    Args:
        texts: 임베딩할 텍스트 리스트

    Returns:
        - dense_vectors: list[list[float]] (각 1024차원)
        - sparse_vectors: list[{"indices": list, "values": list}]
    """
```

---

### 4.2 숫자 파싱 (embedding_service.py)

#### `_parse_number()` - 범위 처리

```python
def _parse_number(value: str, as_int: bool = False) -> float | int | None:
    """
    숫자 문자열을 파싱 (범위 지원)

    Examples:
        "40" → 40.0
        "40.5" → 40.5
        "35~40" → 37.5 (평균)
        "35-40" → 37.5 (평균)
        "N/A" → None
    """
```

**Go 구현:**

```go
func parseNumber(value string, asInt bool) *float64 {
// 범위 패턴: "35~40" 또는 "35-40"
re := regexp.MustCompile(`^(\d+(?:\.\d+)?)[~-](\d+(?:\.\d+)?)$`)
if matches := re.FindStringSubmatch(value); matches != nil {
low, _ := strconv.ParseFloat(matches[1], 64)
high, _ := strconv.ParseFloat(matches[2], 64)
avg := (low + high) / 2
return &avg
}
// 단일 숫자
if num, err := strconv.ParseFloat(value, 64); err == nil {
return &num
}
return nil
}
```

---

### 4.3 임베딩 변환 전략

#### v1: AlcoholEmbeddingInput (기본)

```python
@dataclass
class AlcoholEmbeddingInput:
    id: int
    name_text: str      # "{kor_name} {eng_name} {age}년"
    tags_text: str      # "{tag1_kor} {tag1_eng} {tag2_kor} ..."
    category_text: str  # "{카테고리} {지역} {증류소}"
    full_text: str      # 위 모든 텍스트 합치기
    payload: dict       # id, 이름, 도수, 카테고리 등 메타데이터
```

#### v2: WhiskyEmbeddingStrategy (고급 - 4벡터)

```python
@dataclass
class WhiskyEmbeddingStrategy:
    id: int

    # 4가지 검색 의도별 텍스트
    flavor_semantic_text: str    # 맛/향 (tastingTags + cask + description)
    identity_keyword_text: str   # 브랜드 (이름 + 증류소 + 카테고리)
    origin_context_text: str     # 지역 (region 정보)
    spec_attribute_text: str     # 스펙 (type + abv + age + cask + volume)

    # RAG용 자연어 컨텍스트
    rag_context_text: str

    # 필터링용 메타데이터
    filter_metadata: dict  # {type, abv, age, categoryGroup, region_id, distillery_id, tastingTags}

    # 임베딩 벡터 (각 1024차원)
    flavor_vector: list[float]
    identity_vector: list[float]
    origin_vector: list[float]
    spec_vector: list[float]

    # Sparse 벡터 (키워드 매칭)
    sparse_indices: list[int]
    sparse_values: list[float]
```

---

## 5. API 엔드포인트

### 5.1 술 조회 API (`/alcohols`)

| Method | Path        | Query Params         | Response                | 설명          |
|--------|-------------|----------------------|-------------------------|-------------|
| GET    | `/alcohols` | `start_id`, `end_id` | `list[AlcoholResponse]` | ID 범위로 술 조회 |

**응답 구조:**

```json
{
  "id": 1,
  "kor_name": "맥캘란",
  "eng_name": "Macallan",
  "type": "위스키",
  "abv": "40",
  "volume": "700",
  "age": "12",
  "cask": "셰리 오크",
  "kor_category": "싱글몰트",
  "eng_category": "Single Malt",
  "category_group": "몰트 위스키",
  "description": "...",
  "region": {
    "id": 1,
    "kor_name": "스코틀랜드",
    "eng_name": "Scotland",
    "continent": "유럽",
    "description": "..."
  },
  "distillery": {
    "id": 1,
    "kor_name": "맥캘란 증류소",
    "eng_name": "Macallan Distillery",
    "logo_img_url": "...",
    "description": "..."
  },
  "tasting_tags": [
    {
      "id": 1,
      "kor_name": "달콤한",
      "eng_name": "Sweet",
      "icon": "🍯",
      "description": "..."
    }
  ]
}
```

---

### 5.2 임베딩 조회 API (`/embedding_*`)

| Method | Path            | Query Params   | Response                        | 설명                  |
|--------|-----------------|----------------|---------------------------------|---------------------|
| GET    | `/embedding_v1` | `start`, `end` | `list[AlcoholEmbeddingInput]`   | v1 텍스트 추출 (벡터 X)    |
| GET    | `/embedding_v2` | `start`, `end` | `list[WhiskyEmbeddingStrategy]` | v2 4벡터 + Sparse 임베딩 |

---

### 5.3 Qdrant API (`/qdrant/*`)

| Method | Path                    | Query/Body                   | Response                                 | 설명              |
|--------|-------------------------|------------------------------|------------------------------------------|-----------------|
| POST   | `/qdrant/batch/test`    | `id` (query)                 | `{id, name}`                             | 단건 임베딩 & 저장 테스트 |
| POST   | `/qdrant/batch/start`   | -                            | `{status, total_saved}`                  | 전체 배치 처리        |
| GET    | `/qdrant/search`        | `keyword`, `offset`, `limit` | `{keyword, count, results}`              | 하이브리드 검색 (DBSF) |
| GET    | `/qdrant/search/{type}` | `keyword`, `offset`, `limit` | `{keyword, vector_type, count, results}` | 단일 벡터 타입 검색     |

**Vector Type:**

- `FLAVOR` - 맛/향 기반 검색
- `IDENTITY` - 브랜드/제품명 기반 검색
- `ORIGIN` - 지역/원산지 기반 검색
- `SPEC` - 스펙/특성 기반 검색

---

## 6. 데이터 모델

### 6.1 엔티티 관계도 (ERD)

```
┌──────────────┐     ┌──────────────────────┐     ┌──────────────┐
│   Region     │     │       Alcohol        │     │  Distillery  │
├──────────────┤     ├──────────────────────┤     ├──────────────┤
│ id           │◄────│ region_id (FK)       │     │ id           │
│ kor_name     │     │ distillery_id (FK)───┼────►│ kor_name     │
│ eng_name     │     │ id                   │     │ eng_name     │
│ continent    │     │ kor_name             │     │ logo_img_url │
│ description  │     │ eng_name             │     │ description  │
└──────────────┘     │ type                 │     └──────────────┘
                     │ abv                  │
                     │ volume               │
                     │ age                  │
                     │ cask                 │
                     │ kor_category         │
                     │ eng_category         │
                     │ category_group       │
                     │ image_url            │
                     │ description          │
                     │ create_at/by         │
                     │ last_modify_at/by    │
                     └──────────┬───────────┘
                                │
                                │ M:N
                                ▼
                     ┌──────────────────────┐     ┌──────────────┐
                     │  AlcoholTastingTag   │     │  TastingTag  │
                     ├──────────────────────┤     ├──────────────┤
                     │ id                   │     │ id           │
                     │ alcohol_id (FK)      │────►│ kor_name     │
                     │ tasting_tag_id (FK)  │     │ eng_name     │
                     └──────────────────────┘     │ icon         │
                                                  │ description  │
                                                  └──────────────┘
```

### 6.2 Go Struct 예시

```go
type Region struct {
ID          int64   `gorm:"primaryKey" json:"id"`
KorName     string  `json:"kor_name"`
EngName     string  `json:"eng_name"`
Continent   string  `json:"continent"`
Description *string `json:"description"`
}

type Distillery struct {
ID          int64   `gorm:"primaryKey" json:"id"`
KorName     string  `json:"kor_name"`
EngName     string  `json:"eng_name"`
LogoImgURL  *string `json:"logo_img_url"`
Description *string `json:"description"`
}

type TastingTag struct {
ID          int64   `gorm:"primaryKey" json:"id"`
KorName     string  `json:"kor_name"`
EngName     string  `json:"eng_name"`
Icon        *string `json:"icon"`
Description *string `json:"description"`
}

type Alcohol struct {
ID            int64        `gorm:"primaryKey" json:"id"`
KorName       string       `json:"kor_name"`
EngName       string       `json:"eng_name"`
Type          string       `json:"type"`
ABV           *string      `json:"abv"`
Volume        *string      `json:"volume"`
Age           *string      `json:"age"`
Cask          *string      `json:"cask"`
KorCategory   *string      `json:"kor_category"`
EngCategory   *string      `json:"eng_category"`
CategoryGroup *string      `json:"category_group"`
ImageURL      *string      `json:"image_url"`
Description   *string      `json:"description"`
RegionID      *int64       `json:"region_id"`
DistilleryID  *int64       `json:"distillery_id"`
Region        *Region      `gorm:"foreignKey:RegionID" json:"region"`
Distillery    *Distillery  `gorm:"foreignKey:DistilleryID" json:"distillery"`
TastingTags   []TastingTag `gorm:"many2many:alcohol_tasting_tags" json:"tasting_tags"`
}
```

---

## 7. 임베딩 전략

### 7.1 왜 4개의 벡터가 필요한가?

| 검색 의도      | 벡터 이름      | 텍스트 소스                           | 예시 쿼리                |
|------------|------------|----------------------------------|----------------------|
| **맛/향 검색** | `flavor`   | tastingTags + cask + description | "달콤한 위스키", "스모키한 향"  |
| **브랜드 검색** | `identity` | 이름 + 증류소 + 카테고리                  | "맥캘란", "Macallan 18" |
| **지역 검색**  | `origin`   | region 전체 정보                     | "스코틀랜드 위스키", "아일라"   |
| **스펙 검색**  | `spec`     | type + abv + age + cask + volume | "40도 12년", "셰리캐스크"   |

### 7.2 텍스트 조합 규칙

#### Flavor 텍스트

```python
flavor_parts = []
for tag in alcohol.tasting_tags:
    flavor_parts.append(f"{tag.kor_name} {tag.eng_name}")
    if tag.description:
        flavor_parts.append(tag.description)
if alcohol.cask:
    flavor_parts.append(alcohol.cask)
if alcohol.description:
    flavor_parts.append(alcohol.description)
flavor_semantic_text = " ".join(flavor_parts)
```

#### Identity 텍스트

```python
identity_parts = [alcohol.kor_name, alcohol.eng_name]
if alcohol.distillery:
    identity_parts.extend([alcohol.distillery.kor_name, alcohol.distillery.eng_name])
if alcohol.kor_category:
    identity_parts.extend([alcohol.kor_category, alcohol.eng_category])
identity_keyword_text = " ".join(identity_parts)
```

#### Origin 텍스트

```python
origin_parts = []
if alcohol.region:
    origin_parts.extend([
        alcohol.region.kor_name,
        alcohol.region.eng_name,
        alcohol.region.continent,
        alcohol.region.description or ""
    ])
origin_context_text = " ".join(origin_parts)
```

#### Spec 텍스트

```python
spec_parts = [alcohol.type]
if alcohol.abv:
    spec_parts.append(f"{alcohol.abv}도")
if alcohol.age:
    spec_parts.append(f"{alcohol.age}년")
if alcohol.cask:
    spec_parts.append(alcohol.cask)
if alcohol.volume:
    spec_parts.append(f"{alcohol.volume}ml")
if alcohol.category_group:
    spec_parts.append(alcohol.category_group)
spec_attribute_text = " ".join(spec_parts)
```

### 7.3 RAG 컨텍스트 생성

```python
rag_context = f"{alcohol.kor_name}({alcohol.eng_name})은 "
if alcohol.region:
    rag_context += f"{alcohol.region.kor_name}의 "
if alcohol.distillery:
    rag_context += f"{alcohol.distillery.kor_name}에서 생산된 "
rag_context += f"{alcohol.type}입니다. "
if alcohol.abv:
    rag_context += f"도수는 {alcohol.abv}도이며, "
if alcohol.age:
    rag_context += f"{alcohol.age}년 숙성되었습니다. "
# ... 추가 정보
```

---

## 8. Qdrant 벡터 DB 연동

### 8.1 컬렉션 구조

```python
# 컬렉션명: whisky_v2
# 벡터 크기: 1024 (BGE-m3-ko)

vectors_config = {
    "flavor": VectorParams(size=1024, distance=Distance.COSINE),
    "identity": VectorParams(size=1024, distance=Distance.COSINE),
    "origin": VectorParams(size=1024, distance=Distance.COSINE),
    "spec": VectorParams(size=1024, distance=Distance.COSINE),
}

sparse_vectors_config = {
    "keywords": SparseVectorParams(index=SparseIndexParams(on_disk=False))
}
```

### 8.2 Point 저장 구조

```python
PointStruct(
    id=strategy.id,
    vector={
        "flavor": strategy.flavor_vector,
        "identity": strategy.identity_vector,
        "origin": strategy.origin_vector,
        "spec": strategy.spec_vector,
        "keywords": SparseVector(
            indices=strategy.sparse_indices,
            values=strategy.sparse_values
        )
    },
    payload={
        "rag_context": strategy.rag_context_text,
        **strategy.filter_metadata  # type, abv, age, region_id, etc.
    }
)
```

### 8.3 하이브리드 검색 (DBSF 융합)

```python
# 1. 쿼리 임베딩
dense_vector, sparse = embed_text(keyword)

# 2. Prefetch: 각 벡터별로 상위 N개 후보 추출
prefetch = [
    Prefetch(query=dense_vector, using="flavor", limit=20),
    Prefetch(query=dense_vector, using="identity", limit=20),
    Prefetch(query=dense_vector, using="origin", limit=20),
    Prefetch(query=dense_vector, using="spec", limit=20),
    Prefetch(query=SparseVector(...), using="keywords", limit=20),
]

# 3. DBSF 융합 (Reciprocal Rank Fusion)
results = client.query_points(
    collection_name="whisky_v2",
    prefetch=prefetch,
    query=FusionQuery(fusion=Fusion.DBSF),
    limit=limit,
    offset=offset
)
```

---

## 9. Go 재구현 체크리스트

### Phase 1: 기본 구조

- [ ] 프로젝트 구조 설정 (cmd, internal, pkg)
- [ ] 설정 관리 (환경변수, config 파일)
- [ ] MySQL 연결 (GORM)
- [ ] Qdrant 연결 (qdrant-go)
- [ ] HTTP 서버 설정 (Gin/Fiber)

### Phase 2: 데이터 모델

- [ ] Region 엔티티
- [ ] Distillery 엔티티
- [ ] TastingTag 엔티티
- [ ] AlcoholTastingTag 중간 테이블
- [ ] Alcohol 엔티티 (관계 포함)

### Phase 3: 임베딩 연동

- [ ] 외부 임베딩 API 클라이언트 (Python 서버 호출 또는 Hugging Face API)
- [ ] Dense 벡터 변환
- [ ] Sparse 벡터 변환
- [ ] 배치 처리

### Phase 4: 임베딩 서비스

- [ ] AlcoholEmbeddingInput 변환 (v1)
- [ ] WhiskyEmbeddingStrategy 변환 (v2)
- [ ] 숫자 파싱 (범위 처리)
- [ ] RAG 컨텍스트 생성
- [ ] 필터 메타데이터 추출

### Phase 5: Qdrant 서비스

- [ ] 컬렉션 생성 (4 Named + 1 Sparse)
- [ ] 단건 저장 (upsert)
- [ ] 배치 저장
- [ ] 하이브리드 검색 (DBSF)
- [ ] 단일 벡터 타입 검색

### Phase 6: API 라우터

- [ ] GET /alcohols (범위 조회)
- [ ] GET /embedding_v1 (v1 변환)
- [ ] GET /embedding_v2 (v2 변환)
- [ ] POST /qdrant/batch/test (단건 테스트)
- [ ] POST /qdrant/batch/start (전체 배치)
- [ ] GET /qdrant/search (하이브리드 검색)
- [ ] GET /qdrant/search/{type} (단일 벡터 검색)

### Phase 7: 최적화

- [ ] 연결 풀링 최적화
- [ ] 배치 사이즈 튜닝
- [ ] 에러 핸들링
- [ ] 로깅
- [ ] 헬스체크 엔드포인트

---

## 참고 자료

- **Python 원본 코드**: `/home/hgkim/workspace/embedding`
- **BGE-m3-ko 모델**: https://huggingface.co/dragonkue/BGE-m3-ko
- **Qdrant 문서**: https://qdrant.tech/documentation/
- **qdrant-go 라이브러리**: https://github.com/qdrant/go-client
