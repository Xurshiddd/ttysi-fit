# Security Skills — TTYSI_FIT

Manba: [mukul975/Anthropic-Cybersecurity-Skills](https://github.com/mukul975/Anthropic-Cybersecurity-Skills)
(Apache-2.0, 817 skill, MITRE ATT&CK / NIST CSF ga bog'langan).

To'liq repo katta (59 MB), shuning uchun bu yerga TTYSI_FIT stack'iga
(Go/Gin API, PostgreSQL, Redis, JWT, HEMIS OAuth, Nuxt) **mos 18 ta skill**
tanlab olindi. Har biri `SKILL.md` + `references/` + `scripts/` dan iborat.

## Qanday ishlatiladi

Har bir skill — hujum yoki himoya bo'yicha bosqichma-bosqich yo'riqnoma.
Yangi endpoint yozganda yoki auditda `skills/<nom>/SKILL.md` ni o'qib,
"When to Use" va "Workflow" bo'limlariga amal qiling.

## Skill → CLAUDE.md §17 TOP-50 xaritasi

| Skill | Tegishli hujum (§17.3) |
|-------|------------------------|
| exploiting-sql-injection-vulnerabilities | #1 SQL injection |
| exploiting-api-injection-vulnerabilities | #2–4 injection |
| exploiting-template-injection-vulnerabilities | #12 SSTI |
| exploiting-idor-vulnerabilities | #26 IDOR/BOLA |
| detecting-broken-object-property-level-authorization | #13 mass assignment, BOPLA |
| exploiting-broken-function-level-authorization | #28 BFLA |
| bypassing-authentication-with-forced-browsing | #29 forced browsing |
| exploiting-jwt-algorithm-confusion-attack | #18 JWT alg confusion |
| exploiting-oauth-misconfiguration | #23–24 OAuth CSRF/redirect |
| exploiting-deeplink-vulnerabilities | #9 open redirect / deep link |
| exploiting-race-condition-vulnerabilities | FIT Coin transfer, TOCTOU |
| conducting-api-security-testing | umumiy API test |
| implementing-api-rate-limiting-and-throttling | #15,#16,#40 rate limit |
| implementing-semgrep-for-custom-sast-rules | CI SAST |
| implementing-secret-scanning-with-gitleaks | #33 hardcoded secret |
| performing-sca-dependency-scanning-with-snyk | #48 dependency zaifligi |
| performing-web-application-scanning-with-nikto | #44 header/scan |
| hardening-docker-containers-for-production | infra hardening |

Audit natijasi: `../XAVFSIZLIK_AUDIT.md`
