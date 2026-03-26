Excellent. Now we’re thinking like product + systems engineer, not just coding features.

You’ve accumulated multiple feature directions. Let’s structure and prioritize them properly.

We’ll prioritize based on:

1. Immediate value to you
2. Architectural leverage (what unlocks future power)
3. Complexity vs payoff
4. Performance safety (since this runs in shell hooks)

---

# 🧠 Current State of Recall

You already have:

* ✅ Live recording
* ✅ SQLite persistence
* ✅ Unique constraint
* ✅ Stats (frequency + success rate)
* ✅ FTS search
* ✅ Pretty formatting

So you’re no longer at MVP.
You’re at “foundation built” stage.

Now we optimize intelligence.

---

# 🏆 Priority Strategy

We divide into 3 layers:

---

# 🥇 Tier 1 — Immediate Intelligence (No AI)

These give huge value without complexity.

### 1️⃣ Hybrid Ranking (FTS + frequency + recency)

**Highest priority**

Why?

* Makes search results actually smart
* Uses data you already store
* No new infrastructure
* Big UX improvement

This turns recall from:

> “Search tool”

into:

> “Intelligent recall system”

---

### 2️⃣ Filtering by cwd

Very practical.

You often want:

```bash
recall search grpc --cwd
```

Context-aware recall = high real-world value.

---

### 3️⃣ Success-only filter

Useful for debugging.

```bash
recall search docker --success
```

Avoids noisy failing attempts.

Low complexity, high clarity.

Next logical step:

👉 Add cwd-aware boosting
or
👉 Add success-only filter
or
👉 Add prefix auto-expansion (npx comp → npx comp*)

---

# 🥈 Tier 2 — Smarter Query Layer

### 4️⃣ Fuzzy matching

Nice to have, not critical.

FTS5 prefix search already handles most cases.
Fuzzy adds:

* Levenshtein
* Typo tolerance

Good UX but not foundational.

---

### 5️⃣ Per-day usage analytics

Interesting but not core to recall intelligence.

More “dashboard” than “memory engine”.

---

# 🥉 Tier 3 — Structural Intelligence

### 6️⃣ Session detection

Powerful, but requires:

* Time window grouping
* Possibly heuristics

Great feature — but needs thought.

---

### 7️⃣ Embeddings (semantic search)

High power.
High complexity.
Requires:

* Model
* Embedding storage
* Hybrid ranking tuning

Should only be added after hybrid FTS ranking is excellent.

---

### 8️⃣ Project-aware filtering

Useful.
But depends on:

* Storing cwd properly
* Maybe git root detection

Better after hybrid ranking.

---

# 🎯 Final Recommended Roadmap (Ordered)

## Phase 1 — Smarter Search (No AI)

1️⃣ Improve FTS ranking using bm25()
2️⃣ Add hybrid scoring (frequency + recency)
3️⃣ Build smart ranking engine
4️⃣ Add filtering by cwd
5️⃣ Add success-only filter

This makes recall feel intelligent immediately.

---

## Phase 2 — Behavior Intelligence

6️⃣ Session detection
7️⃣ Project-aware filtering
8️⃣ Per-day usage

---

## Phase 3 — AI Layer

9️⃣ Embeddings (semantic search)
🔟 Hybrid semantic + lexical ranking

---

# 🏆 My Strong Recommendation

We should now do:

> Improve ranking using frequency + recency (Hybrid Search Engine)

Because:

* It uses stats you already computed.
* It upgrades search quality massively.
* It requires zero new infra.
* It prepares foundation for embeddings later.

This is the biggest leverage step.

---

# 🧠 What Hybrid Ranking Will Look Like

Instead of:

```sql
ORDER BY bm25()
```

We compute:

```
final_score =
    (bm25_score * 0.5) +
    (log(frequency) * 0.3) +
    (recency_weight * 0.2)
```

This turns recall into:

> “Search what I most likely meant”

Not just literal text search.

---

# 🚀 So Next Step?

Do we:

A) Implement BM25 weighting + hybrid ranking now
B) Add cwd filtering first
C) Add success-only filter first

If we’re building this strategically, answer is:

👉 Hybrid ranking engine next.

Shall we design the scoring formula properly?
