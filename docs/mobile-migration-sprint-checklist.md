# Mobile Migration Sprint Checklist

## Objetivo

Checklist de execucao da migracao do `self-dojo-mobile` para API do `self-dojo-backend`, com entregas incrementais e criterio de pronto por sprint.

## Premissas Fechadas

- Auth do mobile sera a nativa do backend.
- Upload de avatar sera feito via abstracao no backend (inicialmente com adapter Supabase, trocavel depois).
- Realtime de `watchAcademy` sera substituido por polling no mobile na fase inicial.

## Sprint 0 - Foundation (Arquitetura e Contratos)

- [ ] Definir padrao de versionamento: `/v1`
- [ ] Definir convencao de erro unica (`error`, `details`, `code` opcional)
- [ ] Definir padrao de paginacao/listagem (`page`, `pageSize`, `total`) para listas grandes
- [ ] Definir padrao de filtro para endpoints de busca
- [ ] Definir middleware de auth para mobile + claims necessarios
- [ ] Definir strategy de autorizacao por papel (owner/teacher/instructor/student)
- [ ] Definir interface de storage no backend (`StorageProvider`) + adapter Supabase
- [ ] Criar documento de mapeamento de enums (mobile <-> backend)
- [ ] Revisar `openapi.yaml` atual e separar backlog de rotas novas

Definition of Done:
- [ ] Decisoes acima documentadas em `docs/`
- [ ] Primeira versao de OpenAPI de migracao publicada

## Sprint 1 - Auth e Profile

- [ ] Estender `GET /me` com campos de perfil mobile (`display_name`, `photo_url`, `role`, etc.)
- [ ] Estender `PUT /me` para atualizar os campos necessarios ao mobile
- [ ] Criar `POST /v1/me/avatar` (multipart/form-data)
- [ ] Implementar servico de upload com `StorageProvider`
- [ ] Garantir retorno de URL publica da imagem
- [ ] Atualizar OpenAPI para endpoints de profile
- [ ] Criar testes de integracao de profile/avatar

Definition of Done:
- [ ] Mobile consegue ler/atualizar perfil sem Supabase DB
- [ ] Mobile consegue subir avatar via backend API

## Sprint 2 - Academies e Modalities

- [ ] Criar `POST /v1/academies`
- [ ] Criar `GET /v1/academies?owner=me`
- [ ] Criar `GET /v1/academies/{academyId}`
- [ ] Criar `PATCH /v1/academies/{academyId}`
- [ ] Criar `GET /v1/academies/search`
- [ ] Criar `GET /v1/academies/{academyId}/stats`
- [ ] Criar CRUD base de modalities em `/v1/academies/{academyId}/modalities`
- [ ] Criar `PUT /master` e `PUT /graduation-config`
- [ ] Criar `GET/PUT /teachers` por modalidade
- [ ] Aplicar regras de trial inicial e limites padrao na criacao

Definition of Done:
- [ ] Fluxo de academia no mobile funciona 100% via API
- [ ] Sem chamadas diretas para tabelas `academies`, `academy_modalities`, `belt_configs`

## Sprint 3 - Membership Requests

- [ ] Criar `GET /v1/academies/{academyId}/membership/me`
- [ ] Criar `POST /v1/academies/{academyId}/membership-requests`
- [ ] Criar `DELETE /v1/academies/{academyId}/membership-requests/me`
- [ ] Criar `GET /v1/academies/{academyId}/membership-requests`
- [ ] Criar `POST /approve` e `POST /reject`
- [ ] Garantir idempotencia e bloqueio de solicitacao duplicada
- [ ] Garantir `joined_at` em aprovacao
- [ ] Testar autorizacao (somente staff pode aprovar/rejeitar)

Definition of Done:
- [ ] Fluxo completo de solicitacao de entrada funcionando via API

## Sprint 4 - Students e Enrollment

- [ ] Criar `GET /v1/academies/{academyId}/students`
- [ ] Criar `GET /v1/academies/{academyId}/students/{memberId}`
- [ ] Criar `POST /v1/academies/{academyId}/students/{memberId}/modalities`
- [ ] Criar `PATCH /v1/academies/{academyId}/students/{memberId}/modalities/{studentModalityId}`
- [ ] Criar `DELETE /v1/academies/{academyId}/students/{memberId}/modalities/{studentModalityId}`
- [ ] Incluir retorno com dados de modalidade + historico quando necessario
- [ ] Cobrir regras de autorizacao por academia

Definition of Done:
- [ ] Tela/listagem de alunos e matriculas rodando somente na API

## Sprint 5 - Graduation e Check-in

- [ ] Criar `POST /v1/student-modalities/{studentModalityId}/promotions`
- [ ] Criar `GET /v1/student-modalities/{studentModalityId}/graduation-history`
- [ ] Criar `POST /v1/check-ins`
- [ ] Criar `GET /v1/student-modalities/{studentModalityId}/check-ins`
- [ ] Garantir transacao na promocao (historico + estado atual)
- [ ] Garantir incremento de classes no check-in (sem RPC no mobile)
- [ ] Cobrir testes de regressao de graduacao e check-in

Definition of Done:
- [ ] Fluxos de graduacao e check-in equivalentes ao comportamento Supabase atual

## Sprint 6 - Class Schedules

- [ ] Criar `POST /v1/academies/{academyId}/class-schedules`
- [ ] Criar `GET /v1/academies/{academyId}/class-schedules`
- [ ] Criar `GET /v1/academies/{academyId}/class-schedules/{scheduleId}`
- [ ] Criar `PATCH /v1/academies/{academyId}/class-schedules/{scheduleId}`
- [ ] Criar `DELETE /v1/academies/{academyId}/class-schedules/{scheduleId}`
- [ ] Criar `GET /v1/academies/{academyId}/class-schedules/available-for-checkin`
- [ ] Validar regra de dia da semana + `is_active=true`

Definition of Done:
- [ ] CRUD de horarios e disponibilidade para check-in funcionando via API

## Sprint 7 - Mobile Cutover e Remocao Supabase

- [ ] Criar camada HTTP no mobile para todos os repositorios
- [ ] Migrar `*_repository_supabase.dart` para implementacoes API
- [ ] Remover chamadas `.from(...)`, `.rpc(...)` e `storage` do app
- [ ] Remover dependencia de Supabase DB/Storage do boot do app
- [ ] Rodar smoke test fim a fim em fluxos criticos
- [ ] Atualizar changelog e guias internos

Definition of Done:
- [ ] Mobile sem dependencia funcional de Supabase
- [ ] Toda persistencia mediada por `self-dojo-backend`

## QA e Observabilidade (Transversal)

- [ ] Definir testes de contrato para endpoints mobile-critical
- [ ] Log estruturado com `request_id` em operacoes criticas
- [ ] Metricas basicas por endpoint (latencia/erro)
- [ ] Alertas para falhas em promocao/check-in/upload
- [ ] Plano de rollback por sprint

## Riscos Principais

- Divergencia de modelo de auth entre mobile atual e backend.
- Regras de autorizacao de academia incompletas no primeiro ciclo.
- Regressao em regras transacionais (promocao/check-in).
- Dependencia do provider de storage na entrega de avatar.

## Criterio de Go-Live

- [ ] 100% das features mobile cobertas por endpoints API
- [ ] OpenAPI atualizado e validado
- [ ] Sem acesso direto do mobile ao Supabase
- [ ] Testes de integracao dos fluxos criticos aprovados
- [ ] Monitoramento ativo nos endpoints novos

