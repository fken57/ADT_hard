export type SessionStatus = 'ACTIVE' | 'FINISHED' | 'ABORTED';

export interface TrainingProblem {
  id: string;
  slot: string;
  contestId: string;
  problemId: string;
  problemIndex: string;
  title: string;
  difficulty?: number;
  acceptedAt?: string;
  url: string;
}

export interface TrainingSession {
  id: string;
  atcoderUserId: string;
  startedAt: string;
  durationSeconds: number;
  endedAt?: string;
  status: SessionStatus;
  fallbackLevel: number;
  createdAt: string;
  updatedAt: string;
  problems: TrainingProblem[];
}

export interface SubmissionSync {
  status: 'FRESH' | 'STALE' | 'FAILED' | string;
  lastSuccessfulAt?: string;
  message?: string;
}

export interface SessionResponse {
  session: TrainingSession;
  serverNow: string;
  submissionSync?: SubmissionSync;
}

export interface SessionHistoryResponse {
  sessions: TrainingSession[];
  page: number;
  pageSize: number;
  total: number;
  serverNow: string;
}
