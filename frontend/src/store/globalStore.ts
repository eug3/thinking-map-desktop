import { create } from 'zustand';
import { persist, devtools } from 'zustand/middleware';

interface GlobalStore {
  mapID: string | null;
  loading: boolean;
  error: string | null;
  setMapID: (id: string | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

export const useGlobalStore = create<GlobalStore>()(
  devtools(
    persist(
      (set) => ({
        mapID: null,
        loading: false,
        error: null,
        setMapID: (id) => set({ mapID: id }),
        setLoading: (loading) => set({ loading }),
        setError: (error) => set({ error }),
      }),
      { name: 'global-store' }
    )
  )
);
