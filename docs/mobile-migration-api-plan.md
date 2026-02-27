# Plano de Migracao Mobile (Supabase -> Backend API)

## Objetivo

Migrar o `self-dojo-mobile` de acesso direto ao Supabase (SQL + RPC + Storage) para consumo exclusivo da API do `self-dojo-backend`, sem perda de funcionalidade.

Este documento define:

- inventario funcional atual do mobile;
- gap entre o que existe hoje no backend e o que falta;
- especificacao inicial dos endpoints necessarios;
- ordem recomendada de implementacao/migracao.

## Escopo Atual no Mobile

### Fonte de verdade analisada

- `lib/data/services/supabase_service.dart`
- `lib/data/repositories/*_supabase.dart`
- `supabase/schema.sql`
- `supabase/migrations/*.sql`

### Tabelas/recursos usados pelo app

- `users`
- `academies`
- `academy_modalities`
- `belt_configs`
- `academy_members`
- `modality_teachers`
- `student_modalities`
- `graduation_history`
- `check_ins`
- `class_schedules` (via migration)
- Supabase Storage bucket `avatars`
- RPCs: `increment_student_classes`, `promote_student`

### Features atuais por dominio

- Perfil: buscar/salvar/atualizar perfil, upload de avatar
- Academia: criar, editar, listar do owner, detalhes completos, contadores
- Modalidades: criar/remover modalidade, definir mestre, configurar graduacao, professores por modalidade
- Descoberta de academia: busca por nome/cidade/modalidade
- Vinculo de membro: solicitar entrada, cancelar, listar pendentes, aprovar/rejeitar
- Alunos: listar alunos da academia com modalidades, matricular/desmatricular em modalidade
- Graduacao: promover aluno e historico por modalidade
- Check-in: registrar check-in, listar historico
- Horarios: CRUD de `class_schedules` e busca de horarios disponiveis para check-in

## Cobertura Atual do Backend (Hoje)

Com base no `docs/openapi.yaml` e rotas em `cmd/main.go`, existem:

- `GET /ping`
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/logout`
- `POST /auth/forgot-password`
- `POST /auth/reset-password`
- `GET /me`
- `PUT /me`

### Conclusao de gap

- Apenas **auth/perfil basico** possui endpoints.
- Todo o dominio de academia, membros, modalidades, graduacao, check-in e horarios ainda precisa de API.
- Upload de avatar ainda depende de Supabase Storage no mobile.

## Proposta de Endpoints Necessarios

Padrao sugerido:

- Base prefix: `/v1`
- Autenticacao: `Authorization: Bearer <jwt>`
- IDs: UUID
- Respostas de erro: `{ "error": "...", "details": "..." }`

### 1) Users/Profile

- `GET /v1/me` (ja existe como `/me`; manter compatibilidade)
- `PUT /v1/me` (ja existe como `/me`; estender campos)
- `POST /v1/me/avatar` -> upload/atualizacao de foto de perfil

Campos necessarios em perfil (alem do atual):

- `display_name`
- `photo_url`
- `role`
- `martial_art_type`
- `legacy_belt_id`
- `legacy_degree`
- `legacy_total_classes`
- `legacy_has_aparadores`

### 2) Academies

- `GET /v1/academies/{academyId}` -> detalhes completos da academia (com modalidades e configs)
- `GET /v1/academies?owner=me` -> lista academias do owner
- `POST /v1/academies` -> criar academia
- `PATCH /v1/academies/{academyId}` -> atualizar academia
- `GET /v1/academies/search?q=&city=&modality=` -> busca publica/filtrada
- `GET /v1/academies/{academyId}/stats` -> contadores (membros aprovados, pendentes)

### 3) Academy Modalities

- `POST /v1/academies/{academyId}/modalities`
- `GET /v1/academies/{academyId}/modalities`
- `PATCH /v1/academies/{academyId}/modalities/{modalityId}`
- `DELETE /v1/academies/{academyId}/modalities/{modalityId}`

Sub-recursos:

- `PUT /v1/academies/{academyId}/modalities/{modalityId}/master`
- `PUT /v1/academies/{academyId}/modalities/{modalityId}/graduation-config`
- `GET /v1/academies/{academyId}/modalities/{modalityId}/teachers`
- `PUT /v1/academies/{academyId}/modalities/{modalityId}/teachers` (substitui lista)

### 4) Membership (solicitacoes e membros)

- `GET /v1/academies/{academyId}/membership/me` -> status do usuario logado nessa academia
- `POST /v1/academies/{academyId}/membership-requests` -> solicitar entrada
- `DELETE /v1/academies/{academyId}/membership-requests/me` -> cancelar solicitacao pendente
- `GET /v1/academies/{academyId}/membership-requests?status=pending`
- `POST /v1/academies/{academyId}/membership-requests/{memberId}/approve`
- `POST /v1/academies/{academyId}/membership-requests/{memberId}/reject`

### 5) Students

- `GET /v1/academies/{academyId}/students` -> lista alunos aprovados + modalidades
- `GET /v1/academies/{academyId}/students/{memberId}` -> detalhe do aluno
- `POST /v1/academies/{academyId}/students/{memberId}/modalities` -> matricular em modalidade
- `DELETE /v1/academies/{academyId}/students/{memberId}/modalities/{studentModalityId}` -> desmatricular
- `PATCH /v1/academies/{academyId}/students/{memberId}/modalities/{studentModalityId}` -> atualizacoes gerais (classes, teacher, etc)

### 6) Graduation

- `POST /v1/student-modalities/{studentModalityId}/promotions`
  - Equivale ao RPC `promote_student` (transacional)
  - Atualiza faixa/grau e grava `graduation_history`
- `GET /v1/student-modalities/{studentModalityId}/graduation-history`

### 7) Check-ins

- `POST /v1/check-ins`
  - Cria check-in
  - Deve incrementar contadores de aulas (equivalente ao RPC `increment_student_classes`)
- `GET /v1/student-modalities/{studentModalityId}/check-ins?startDate=&endDate=`

### 8) Class Schedules

- `POST /v1/academies/{academyId}/class-schedules`
- `GET /v1/academies/{academyId}/class-schedules?modalityId=&dayOfWeek=&isActive=`
- `GET /v1/academies/{academyId}/class-schedules/{scheduleId}`
- `PATCH /v1/academies/{academyId}/class-schedules/{scheduleId}`
- `DELETE /v1/academies/{academyId}/class-schedules/{scheduleId}`
- `GET /v1/academies/{academyId}/class-schedules/available-for-checkin`

## Contratos Minimos por Entidade (DTOs)

### Academy

- `id`, `owner_id`, `name`, `description`, `logo_url`
- `address`, `city`, `state`, `phone`, `email`, `website`
- `subscription_plan`, `subscription_started_at`, `subscription_ends_at`
- `max_students`, `max_teachers`, `max_modalities`, `is_active`
- `academy_modalities[]`

### AcademyModality

- `id`, `academy_id`, `martial_art_type`, `master_id`
- `use_default_graduation`, `graduation_updated_at`, `is_active`
- `belt_configs[]`
- `teachers[]` (quando solicitado expandido)

### AcademyMember

- `id`, `academy_id`, `user_id`, `role`, `status`
- `payment_status`, `payment_due_date`
- `joined_at`, `approved_at`, `approved_by`
- `user` (expandido): `display_name`, `email`, `photo_url`

### StudentModality

- `id`, `member_id`, `modality_id`
- `assigned_teacher_id`
- `belt_id`, `degree`, `promotion_date`
- `total_classes`, `classes_at_current_belt`
- `enrolled_at`
- `graduation_history[]` (expandido opcional)

### CheckIn

- `id`, `student_modality_id`, `class_schedule_id`
- `checked_in_at`, `checked_in_by`
- `class_type`, `duration_minutes`, `notes`

### ClassSchedule

- `id`, `academy_id`, `modality_id`, `instructor_id`
- `day_of_week`, `start_time`, `end_time`
- `class_type`, `is_active`, `max_students`, `notes`

## Regras de Negocio Criticas (manter comportamento atual)

- Criacao de academia inicia `subscription_plan=trial` com janela de 7 dias e limites padrao.
- Solicitacao de membro nao pode duplicar pendente/aprovada para mesmo usuario+academia.
- `approve` de membro deve definir `status=approved` e `joined_at`.
- Promocao de aluno deve ser transacional (historico + estado atual).
- Check-in deve incrementar contadores de classes da modalidade do aluno.
- Busca de academias deve filtrar academias inativas.
- Disponibilidade de horarios para check-in considera dia atual e `is_active=true`.

## Matriz de Migracao (feature -> endpoint)

- Perfil -> `/v1/me`, `/v1/me/avatar`
- Academia owner -> `/v1/academies`, `/v1/academies?owner=me`, `/v1/academies/{id}`
- Busca de academias -> `/v1/academies/search`
- Solicitacoes de vinculo -> `/v1/academies/{id}/membership-requests`*
- Alunos da academia -> `/v1/academies/{id}/students`*
- Modalidades e graduacao -> `/v1/academies/{id}/modalities`*, `/v1/student-modalities/{id}/promotions`
- Check-in -> `/v1/check-ins`, `/v1/student-modalities/{id}/check-ins`
- Horarios -> `/v1/academies/{id}/class-schedules*`

## Ordem Recomendada de Implementacao

1. **Base de identidade/perfil**
  - consolidar `/me` com campos necessarios e avatar
2. **Academias + modalidades (read/write)**
3. **Membership requests**
4. **Students + enrollment**
5. **Graduation + check-in (operacoes transacionais)**
6. **Class schedules**
7. **Atualizar OpenAPI e gerar cliente HTTP para mobile**
8. **Trocar repositorios Supabase por repositorios API no mobile**
9. **Remover dependencia Supabase do app (DB/Storage)**

## Criterios de Pronto para Migracao

- Todos os fluxos acima funcionando somente via backend API.
- Nenhuma chamada a `.from(...)`, `.rpc(...)` ou `storage` no mobile.
- Cobertura de testes de integracao para:
  - aprovacao/rejeicao de membro;
  - promocao de aluno;
  - check-in com incremento de classes;
  - CRUD de horarios.
- `docs/openapi.yaml` refletindo 100% dos endpoints expostos para mobile.

## Duvidas/Decisoes Pendentes (bloqueantes)

- Autenticacao final do mobile sera:
  - manter Firebase e trocar token no backend, ou
  - migrar para auth nativa do backend (`/auth/login` e `/auth/register`)? vamos usar a nativa do backend
- Upload de avatar ficara:
  - storage proprio no backend (S3/local), ou
  - backend como proxy para Supabase Storage temporariamente? vamos fazer uma abstração para usar o supabase podendo trocar para outro quando necessario
- Necessidade de realtime:
  - hoje existe `watchAcademy` no Supabase; definir se sera substituido por polling ou websocket.

## Artefatos Complementares

- Checklist operacional por sprint: `docs/mobile-migration-sprint-checklist.md`
- Template de especificacao OpenAPI da migracao: `docs/openapi-mobile-template.yaml`

