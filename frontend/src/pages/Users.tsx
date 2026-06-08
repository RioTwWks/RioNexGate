import { useEffect, useState } from 'react';
import api from '../services/api';

export const Users = () => {
  const [users, setUsers] = useState([]);

  const fetchUsers = async () => {
    const res = await api.get('/users');
    setUsers(res.data);
  };

  const addUser = async () => {
    await api.post('/users', { email: 'new@example.com', traffic_gb: 50 });
    fetchUsers();
  };

  return (
    <div>
      <button onClick={addUser}>Add User</button>
      <table>...</table>
    </div>
  );
};