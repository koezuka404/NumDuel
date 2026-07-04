type Props = {
  value: string;
  onChange: (value: string) => void;
  onSubmit: () => void;
  disabled?: boolean;
  submitLabel?: string;
  maxDigits?: number;
};

export default function NumericKeypad({
  value,
  onChange,
  onSubmit,
  disabled = false,
  submitLabel = '送信',
  maxDigits = 4,
}: Props) {
  const appendDigit = (digit: string) => {
    if (disabled || value.length >= maxDigits) {
      return;
    }
    if (value.includes(digit)) {
      return;
    }
    onChange(value + digit);
  };

  const backspace = () => {
    if (disabled) {
      return;
    }
    onChange(value.slice(0, -1));
  };

  const keys = ['1', '2', '3', '4', '5', '6', '7', '8', '9'];

  return (
    <div className="game-keypad">
      <div className="game-keypad__display" aria-live="polite">
        {value.padEnd(maxDigits, '·').split('').map((char, index) => (
          <span key={index} className={`game-keypad__slot ${char !== '·' ? 'game-keypad__slot--filled' : ''}`}>
            {char === '·' ? '' : char}
          </span>
        ))}
      </div>
      <div className="game-keypad__grid">
        {keys.map((key) => (
          <button
            key={key}
            type="button"
            className="game-keypad__key"
            disabled={disabled || value.includes(key)}
            onClick={() => appendDigit(key)}
          >
            {key}
          </button>
        ))}
        <button type="button" className="game-keypad__key game-keypad__key--action" disabled={disabled} onClick={backspace}>
          ⌫
        </button>
        <button type="button" className="game-keypad__key" disabled={disabled || value.includes('0')} onClick={() => appendDigit('0')}>
          0
        </button>
        <button
          type="button"
          className="game-keypad__key game-keypad__key--submit"
          disabled={disabled || value.length !== maxDigits}
          onClick={onSubmit}
        >
          {submitLabel}
        </button>
      </div>
    </div>
  );
}
