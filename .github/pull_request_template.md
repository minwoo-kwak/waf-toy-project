# Pull Request

## 📋 요약 (Summary)
<!-- 이 PR에서 구현한 기능이나 수정사항을 간단히 설명해주세요 -->

## 🛠️ 변경사항 (Changes)
<!-- 구체적인 변경 내용을 체크리스트로 작성해주세요 -->
- [ ] 새로운 기능 추가
- [ ] 버그 수정
- [ ] 코드 리팩토링
- [ ] 문서 업데이트
- [ ] 테스트 추가/수정
- [ ] 설정 파일 수정

## 🎯 관련 이슈 (Related Issues)
<!-- 관련된 이슈 번호를 작성해주세요 -->
Closes #이슈번호

## 🧪 테스트 (Testing)
<!-- 어떤 테스트를 수행했는지 작성해주세요 -->
### 테스트 환경
- [ ] 로컬 개발 환경
- [ ] Docker Desktop Kubernetes
- [ ] WAF 보안 테스트

### 테스트 결과
```bash
# 테스트 명령어와 결과를 작성해주세요
```

## 📸 스크린샷 또는 데모 (Screenshots/Demo)
<!-- UI 변경사항이 있다면 스크린샷을 첨부해주세요 -->

## ⚠️ 주의사항 (Notes)
<!-- 리뷰어가 특별히 확인해야 할 사항이나 주의점 -->
- 
- 

## 📝 체크리스트 (Checklist)
### 개발자 체크
- [ ] 코드 스타일 가이드를 준수했습니다
- [ ] 단위 테스트를 작성/업데이트했습니다
- [ ] 문서를 업데이트했습니다 (필요시)
- [ ] 보안 취약점을 검토했습니다
- [ ] 성능에 미치는 영향을 고려했습니다

### WAF 프로젝트 특화 체크
- [ ] ModSecurity 설정이 올바르게 적용됩니다
- [ ] OWASP CRS 룰이 정상 작동합니다
- [ ] Kubernetes 매니페스트가 유효합니다
- [ ] Docker 이미지가 정상 빌드됩니다
- [ ] 보안 테스트를 수행했습니다 (SQL Injection, XSS 등)

## 🔄 배포 (Deployment)
### 배포 단계
- [ ] 개발 환경에서 테스트 완료
- [ ] Docker 이미지 빌드 확인
- [ ] Kubernetes 배포 확인
- [ ] WAF 로그 정상 동작 확인

### 배포 명령어
```bash
# 배포에 필요한 명령어를 작성해주세요
./scripts/deploy-k8s.sh
```

## 👥 리뷰어 (Reviewers)
<!-- 리뷰를 요청할 팀원을 태그해주세요 -->
@리뷰어명

## 📚 참고 자료 (References)
<!-- 관련 문서나 참고 자료 링크 -->
- [ModSecurity Documentation](https://github.com/SpiderLabs/ModSecurity/wiki)
- [OWASP CRS](https://coreruleset.org/)

---
**📌 이 PR은 WAF SaaS Platform 프로젝트의 [N]주차 과제입니다.**