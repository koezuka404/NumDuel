import { useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { ApiError, apiData } from '../api/client';
import { useAuth } from './useAuth';
import { useGameState } from './useGameState';
import { useToast } from './useToast';
import { useWebSocket } from './useWebSocket';
import { validateFourDigits } from '../lib/validation';
import type { GameOverData, GameStateDTO } from '../types/dto';

const DEFAULT_TURN_SECONDS = 30;
const DEFAULT_SECRET_SECONDS = 60;

export function useGamePage() {
  const { id: gameId = '' } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { showToast } = useToast();
  const { send, subscribe, connected } = useWebSocket();
  const { state, dispatch } = useGameState();

  const [secretInput, setSecretInput] = useState('');
  const [guessInput, setGuessInput] = useState('');
  const [inputError, setInputError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [reconnectBanner, setReconnectBanner] = useState('');
  const [timerMax, setTimerMax] = useState(DEFAULT_TURN_SECONDS);
  const pendingGuessRef = useRef<string | null>(null);

  useEffect(() => {
    if (!gameId || !user) {
      return;
    }
    dispatch({ type: 'LOADING' });
    apiData<GameStateDTO>(`/games/${gameId}`)
      .then((data) => {
        dispatch({ type: 'SET_STATE', state: data, userId: user.id });
        setTimerMax(data.status === 'WAITING_SECRET' ? DEFAULT_SECRET_SECONDS : DEFAULT_TURN_SECONDS);
        if (data.status === 'FINISHED') {
          dispatch({
            type: 'GAME_OVER',
            data: { gameId, reason: 'guess_win' },
          });
        }
      })
      .catch((err) => {
        if (err instanceof ApiError) {
          showToast(err.message, 'error');
          if (err.code === 'not_found' || err.code === 'forbidden') {
            navigate('/matching');
          }
        }
      });
  }, [gameId, user, dispatch, navigate, showToast]);

  useEffect(() => {
    if (!user) {
      return;
    }
    return subscribe((msg) => {
      if (msg.type === 'GAME_STATE_SYNC' && msg.data) {
        const data = msg.data as unknown as GameStateDTO;
        dispatch({ type: 'SET_STATE', state: data, userId: user.id });
        if (data.remainingSeconds > 0) {
          setTimerMax((prev) => Math.max(prev, data.remainingSeconds));
        }
      }
      if (msg.type === 'TURN_CHANGED') {
        dispatch({ type: 'TURN_CHANGED', data: msg.data ?? {} });
        const remaining = Number(msg.data?.remainingSeconds ?? DEFAULT_TURN_SECONDS);
        setTimerMax(Math.max(remaining, 1));
      }
      if (msg.type === 'GUESS_RESULT') {
        dispatch({
          type: 'GUESS_RESULT',
          data: msg.data ?? {},
          userId: user.id,
          pendingGuess: pendingGuessRef.current ?? undefined,
        });
        pendingGuessRef.current = null;
      }
      if (msg.type === 'GAME_OVER') {
        dispatch({ type: 'GAME_OVER', data: msg.data as unknown as GameOverData });
      }
      if (msg.type === 'OPPONENT_STATUS') {
        dispatch({ type: 'OPPONENT_STATUS', connected: Boolean(msg.data?.connected) });
      }
      if (msg.type === 'ERROR') {
        const code = String(msg.data?.code ?? '');
        const message = String(msg.data?.message ?? 'エラーが発生しました');
        if (code === 'not_found' || code === 'forbidden') {
          showToast(message, 'error');
          navigate('/matching');
          return;
        }
        if (code === 'game_already_finished') {
          navigate('/matching');
          return;
        }
        showToast(message, 'error');
      }
    });
  }, [subscribe, user, dispatch, navigate, showToast]);

  useEffect(() => {
    if (!connected && user?.role === 'user') {
      setReconnectBanner('再接続中…');
    } else if (connected && reconnectBanner) {
      setReconnectBanner('接続が復旧しました');
      const timer = window.setTimeout(() => setReconnectBanner(''), 3000);
      return () => window.clearTimeout(timer);
    }
  }, [connected, user?.role, reconnectBanner]);

  useEffect(() => {
    if (connected && gameId) {
      send({ type: 'SYNC_REQUEST', gameId });
    }
  }, [connected, gameId, send]);

  const isMyTurn = user?.id === state.currentTurnPlayerID;
  const isSecretPhase = state.status === 'WAITING_SECRET';
  const isPlaying = state.status === 'IN_PROGRESS';
  const inputValue = isSecretPhase ? secretInput : guessInput;
  const setInputValue = isSecretPhase ? setSecretInput : setGuessInput;
  const inputDisabled = isSecretPhase ? state.secretSubmitted || submitting : !isMyTurn || submitting;

  const submitCurrentInput = () => {
    const err = validateFourDigits(inputValue);
    if (err) {
      setInputError(err);
      return;
    }
    setInputError('');
    setSubmitting(true);
    if (isSecretPhase) {
      send({ type: 'SET_SECRET', gameId, secretNumber: inputValue });
      dispatch({ type: 'SECRET_SUBMITTED' });
      setSecretInput('');
    } else {
      pendingGuessRef.current = inputValue;
      send({ type: 'GUESS', gameId, guessNumber: inputValue });
      setGuessInput('');
    }
    setSubmitting(false);
  };

  const closeResult = () => {
    dispatch({ type: 'CLEAR_GAME_OVER' });
    navigate('/matching');
  };

  return {
    gameId,
    user,
    state,
    reconnectBanner,
    timerMax,
    isMyTurn,
    isSecretPhase,
    isPlaying,
    inputValue,
    setInputValue,
    inputError,
    setInputError,
    inputDisabled,
    submitCurrentInput,
    closeResult,
  };
}
