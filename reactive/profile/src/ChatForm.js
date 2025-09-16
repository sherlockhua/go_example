import { useState } from "react";

import { people } from "./data.js";

export default function ChatForm() {
  const [to, setTo] = useState("");
  const [message, setMessage] = useState("");
  const handleSubmit = (e) => {
    e.preventDefault();
    console.log("Sending message to:", to);
    alert(`Sending message "${message}" to ${to}`);
  };
  return (
    <form onSubmit={handleSubmit}>
      <li>To: {to} </li>
      <select value={to} onChange={(e) => setTo(e.target.value)}>
        {people.map((person) => (
          <option key={person.id} value={person.name}>
            {person.name}
          </option>
        ))}
      </select>
      <div>
        <input
          type="text"
          placeholder="Type a message"
          value={message}
          onChange={(e) => setMessage(e.target.value)}
        />
      </div>
      <button type="submit">Send</button>
    </form>
  );
}
