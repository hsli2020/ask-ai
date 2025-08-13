import React, { useState } from 'react';

const CounterComponent = () => {
  const [counter, setCounter] = useState(0);
  const [inputText, setInputText] = useState('');
  const [showInput, setShowInput] = useState(false);

  const handleClick = () => {
    setCounter(counter + 1);
  };

  const handleInputChange = (event) => {
    setInputText(event.target.value);
  };

  const handleCheckboxChange = (event) => {
    setShowInput(event.target.checked);
  };

  return (
    <div>
      <p>Counter: {counter}</p>
      <button onClick={handleClick}>Increase Counter</button>
      <input type="checkbox" onChange={handleCheckboxChange} />
      <label htmlFor="hide-show-input">Hide/Show Input Field</label>
      {showInput && (
        <>
          <input type="text" value={inputText} onChange={handleInputChange} />
          <p>{inputText}</p>
        </>
      )}
    </div>
  );
};
