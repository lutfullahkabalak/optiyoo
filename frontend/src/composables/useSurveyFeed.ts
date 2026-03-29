import { nextTick, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { absoluteApiUrl, resolveAvatarBackground, usernameInitial } from '../utils/avatar'

export const surveyFeedApiBase = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export type SurveyListApply = (list: any[]) => void

export function useSurveyFeed(opts: {
  onConflictReload: (apply: SurveyListApply) => Promise<void>
}) {
  const router = useRouter()
  const authStore = useAuthStore()

  const surveys = ref<any[]>([])
  const givenAnswers = ref<Record<string, string>>({})
  const copiedSurveyId = ref<string | null>(null)
  const animatedPercents = ref<Record<string, number>>({})
  const openMenuSurveyId = ref<string | null>(null)
  const activeQuestionIndexes = ref<Record<string, number>>({})
  const submittingState = ref<Record<string, boolean>>({})
  const completedSurveys = ref<Set<string>>(
    new Set(JSON.parse(localStorage.getItem('completed_polls_' + authStore.user?.id) || '[]'))
  )
  const lockedQuestionIdsBySurvey = ref<Record<string, Record<string, true>>>({})

  const isSurveyFullyAnswered = (survey: any, answers: Record<string, string>) => {
    const questions = survey?.questions || []
    if (questions.length === 0) return false
    return questions.every((q: any) => q.id != null && Boolean(answers[String(q.id)]))
  }

  const persistCompletedSurveys = () => {
    const uid = authStore.user?.id
    if (!uid) return
    localStorage.setItem('completed_polls_' + uid, JSON.stringify(Array.from(completedSurveys.value)))
  }

  const applySurveyListPayload: SurveyListApply = (list: any[]) => {
    surveys.value = list
    list.forEach((s: any) => {
      if (activeQuestionIndexes.value[s.id] === undefined) {
        activeQuestionIndexes.value[s.id] = 0
      }
      const ua = s.user_answers as Record<string, string> | undefined
      if (ua && typeof ua === 'object' && Object.keys(ua).length > 0) {
        const prevLocks = lockedQuestionIdsBySurvey.value[s.id] || {}
        const merged: Record<string, true> = { ...prevLocks }
        const nextAnswers = { ...givenAnswers.value }
        for (const [qid, optId] of Object.entries(ua)) {
          const k = String(qid)
          nextAnswers[k] = String(optId)
          merged[k] = true
        }
        givenAnswers.value = nextAnswers
        lockedQuestionIdsBySurvey.value = {
          ...lockedQuestionIdsBySurvey.value,
          [s.id]: merged
        }
        if (isSurveyFullyAnswered(s, nextAnswers)) {
          completedSurveys.value.add(s.id)
        } else {
          completedSurveys.value.delete(s.id)
        }
      } else if (s.user_answer) {
        const firstQuestionId = s?.questions?.[0]?.id
        if (firstQuestionId) {
          const fid = String(firstQuestionId)
          const nextAnswers = { ...givenAnswers.value, [fid]: String(s.user_answer) }
          givenAnswers.value = nextAnswers
          const prevLocks = lockedQuestionIdsBySurvey.value[s.id] || {}
          lockedQuestionIdsBySurvey.value = {
            ...lockedQuestionIdsBySurvey.value,
            [s.id]: { ...prevLocks, [fid]: true }
          }
          if (isSurveyFullyAnswered(s, nextAnswers)) {
            completedSurveys.value.add(s.id)
          } else {
            completedSurveys.value.delete(s.id)
          }
        }
      }
      if (completedSurveys.value.has(s.id)) {
        setSurveyPercentsInstant(s, getCurrentQuestion(s))
      }
    })
    persistCompletedSurveys()
  }

  const copyLink = async (surveyId: string) => {
    const link = `${window.location.origin}/s/${surveyId}`
    await navigator.clipboard.writeText(link)
    copiedSurveyId.value = surveyId
    setTimeout(() => (copiedSurveyId.value = null), 1500)
  }

  const toggleSurveyMenu = (surveyId: string) => {
    openMenuSurveyId.value = openMenuSurveyId.value === surveyId ? null : surveyId
  }

  const closeSurveyMenu = () => {
    openMenuSurveyId.value = null
  }

  const openSurveyDetail = (surveyId: string) => {
    closeSurveyMenu()
    router.push(`/s/${surveyId}`)
  }

  const getTotalVotes = (q: any) => {
    if (!q || !q.options) return 0
    return q.options.reduce((sum: number, o: any) => sum + (o.vote_count || 0), 0)
  }

  const getOptionPercent = (q: any, opt: any) => {
    const totalVotes = getTotalVotes(q)
    if (totalVotes <= 0) return 0
    return Math.round(((opt.vote_count || 0) / totalVotes) * 100)
  }

  const getOptionAnimKey = (surveyId: string, optionId: string) => `${surveyId}:${optionId}`

  const getCurrentQuestion = (survey: any) => {
    const questions = survey?.questions || []
    if (questions.length === 0) return null
    const activeIndex = activeQuestionIndexes.value[survey.id] ?? 0
    return questions[Math.min(Math.max(activeIndex, 0), questions.length - 1)]
  }

  const isCurrentQuestionAnswerLocked = (survey: any) => {
    const q = getCurrentQuestion(survey)
    if (q?.id == null || completedSurveys.value.has(survey.id)) return false
    return Boolean(lockedQuestionIdsBySurvey.value[survey.id]?.[String(q.id)])
  }

  const canGoPrevQuestion = (survey: any) => (activeQuestionIndexes.value[survey.id] ?? 0) > 0

  const canGoNextQuestion = (survey: any) => {
    const questions = survey?.questions || []
    return (activeQuestionIndexes.value[survey.id] ?? 0) < questions.length - 1
  }

  const goToPrevQuestion = (survey: any) => {
    if (!canGoPrevQuestion(survey)) return
    activeQuestionIndexes.value[survey.id] = (activeQuestionIndexes.value[survey.id] ?? 0) - 1
  }

  const goToNextQuestion = (survey: any) => {
    if (!canGoNextQuestion(survey)) return
    activeQuestionIndexes.value[survey.id] = (activeQuestionIndexes.value[survey.id] ?? 0) + 1
  }

  const getAnimatedPercent = (survey: any, q: any, opt: any) => {
    const key = getOptionAnimKey(survey.id, opt.id)
    return animatedPercents.value[key] ?? getOptionPercent(q, opt)
  }

  const setSurveyPercentsInstant = (survey: any, q: any) => {
    if (!q?.options) return
    q.options.forEach((opt: any) => {
      const key = getOptionAnimKey(survey.id, opt.id)
      animatedPercents.value[key] = getOptionPercent(q, opt)
    })
  }

  const animateSurveyPercents = (survey: any, q: any) => {
    if (!q?.options) return
    const targets: Array<{ key: string; target: number }> = q.options.map((opt: any) => ({
      key: getOptionAnimKey(survey.id, opt.id),
      target: getOptionPercent(q, opt)
    }))
    targets.forEach(({ key }) => {
      animatedPercents.value[key] = 0
    })
    const duration = 275
    const start = performance.now()
    const tick = (now: number) => {
      const progress = Math.min((now - start) / duration, 1)
      const eased = 1 - Math.pow(1 - progress, 3)
      targets.forEach(({ key, target }) => {
        animatedPercents.value[key] = Math.round(target * eased)
      })
      if (progress < 1) requestAnimationFrame(tick)
    }
    requestAnimationFrame(tick)
  }

  const revertQuestionSelection = (survey: any, questionId: string) => {
    const qid = String(questionId)
    const prevLocks = { ...(lockedQuestionIdsBySurvey.value[survey.id] || {}) }
    delete prevLocks[qid]
    lockedQuestionIdsBySurvey.value = {
      ...lockedQuestionIdsBySurvey.value,
      [survey.id]: prevLocks
    }
    const next = { ...givenAnswers.value }
    delete next[qid]
    givenAnswers.value = next
  }

  const handleQuestionSelect = async (survey: any, payload: { questionId: string; optionId: string }) => {
    if (completedSurveys.value.has(survey.id)) return
    const qid = String(payload.questionId)
    if (lockedQuestionIdsBySurvey.value[survey.id]?.[qid]) return
    givenAnswers.value = { ...givenAnswers.value, [qid]: String(payload.optionId) }
    const prev = lockedQuestionIdsBySurvey.value[survey.id] || {}
    lockedQuestionIdsBySurvey.value = {
      ...lockedQuestionIdsBySurvey.value,
      [survey.id]: { ...prev, [qid]: true }
    }
    await submitSingleAnswer(survey, qid, String(payload.optionId))
  }

  const submitSingleAnswer = async (survey: any, questionId: string, optionId: string) => {
    if (submittingState.value[survey.id]) return
    submittingState.value[survey.id] = true
    const payload = {
      answers: [{ question_id: questionId, value: optionId }]
    }
    try {
      const res = await fetch(`${surveyFeedApiBase}/api/surveys/${survey.id}/answers`, {
        method: 'POST',
        headers: authStore.authHeadersJson(),
        body: JSON.stringify(payload)
      })
      if (res.ok) {
        const answeredQ = survey.questions?.find((q: any) => String(q.id) === questionId)
        const opt = answeredQ?.options?.find((o: any) => String(o.id) === optionId)
        if (opt) opt.vote_count = (opt.vote_count || 0) + 1
        if (isSurveyFullyAnswered(survey, givenAnswers.value)) {
          completedSurveys.value.add(survey.id)
          persistCompletedSurveys()
        }
        await nextTick()
        if (answeredQ) animateSurveyPercents(survey, answeredQ)
      } else {
        const text = await res.text()
        if (res.status === 403) {
          await opts.onConflictReload(applySurveyListPayload)
        } else {
          revertQuestionSelection(survey, questionId)
          alert('Oylama İptali: ' + text)
        }
      }
    } catch (e) {
      console.error(e)
      revertQuestionSelection(survey, questionId)
    } finally {
      submittingState.value[survey.id] = false
    }
  }

  const timeAgo = (dateStr: string) => {
    if (!dateStr) return ''
    const parseCreatedAtMs = (value: string) => {
      const now = Date.now()
      const primary = new Date(value).getTime()
      if (!Number.isNaN(primary)) {
        if (primary > now && value.endsWith('Z')) {
          const withoutZ = value.slice(0, -1)
          const localFallback = new Date(withoutZ).getTime()
          if (!Number.isNaN(localFallback) && localFallback <= now) return localFallback
        }
        return primary
      }
      const match = value.match(
        /^(\d{4})-(\d{2})-(\d{2})[ T](\d{2}):(\d{2}):(\d{2})(?:\.(\d+))?$/
      )
      if (match) {
        const [, y, mo, d, h, mi, s, frac = '0'] = match
        const ms = Number(frac.padEnd(3, '0').slice(0, 3))
        return new Date(
          Number(y),
          Number(mo) - 1,
          Number(d),
          Number(h),
          Number(mi),
          Number(s),
          ms
        ).getTime()
      }
      return now
    }
    const createdAtMs = parseCreatedAtMs(dateStr)
    const rawDiff = Math.floor((Date.now() - createdAtMs) / 1000)
    const diff = Math.max(0, rawDiff)
    if (diff < 60) return 'şimdi'
    if (diff < 3600) return `${Math.floor(diff / 60)}dk`
    if (diff < 86400) return `${Math.floor(diff / 3600)}sa`
    return `${Math.floor(diff / 86400)}g`
  }

  const getCreatorName = (survey: any) => survey.creator_name || 'Optiyoo'

  const getCreatorHandle = (survey: any) => {
    const h = String(survey.creator_username || '').trim()
    if (h) return `@${h}`
    const name = getCreatorName(survey)
    return `@${name.toLowerCase().replace(/\s+/g, '')}`
  }

  const creatorAvatarImageUrl = (survey: any) =>
    absoluteApiUrl(surveyFeedApiBase, survey.creator_avatar_url as string | undefined)

  const creatorAvatarBg = (survey: any) =>
    resolveAvatarBackground(survey.creator_avatar_color, survey.creator_id || survey.id)

  const creatorAvatarLetter = (survey: any) => usernameInitial(survey.creator_username)

  return {
    surveys,
    givenAnswers,
    copiedSurveyId,
    animatedPercents,
    openMenuSurveyId,
    activeQuestionIndexes,
    submittingState,
    completedSurveys,
    lockedQuestionIdsBySurvey,
    applySurveyListPayload,
    copyLink,
    toggleSurveyMenu,
    closeSurveyMenu,
    openSurveyDetail,
    getTotalVotes,
    getOptionPercent,
    getOptionAnimKey,
    getCurrentQuestion,
    isCurrentQuestionAnswerLocked,
    canGoPrevQuestion,
    canGoNextQuestion,
    goToPrevQuestion,
    goToNextQuestion,
    getAnimatedPercent,
    setSurveyPercentsInstant,
    animateSurveyPercents,
    handleQuestionSelect,
    submitSingleAnswer,
    timeAgo,
    getCreatorName,
    getCreatorHandle,
    creatorAvatarImageUrl,
    creatorAvatarBg,
    creatorAvatarLetter
  }
}
