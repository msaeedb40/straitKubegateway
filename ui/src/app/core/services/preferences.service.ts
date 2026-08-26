import { Injectable, signal, inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';

export interface UserPreferences {
  theme: 'dark' | 'light';
  tableDensity: 'compact' | 'comfortable';
  telemetryRefreshInterval: number;
  enableSounds: boolean;
  d3SimulationEnabled: boolean;
}

const DEFAULT_PREFS: UserPreferences = {
  theme: 'dark',
  tableDensity: 'comfortable',
  telemetryRefreshInterval: 5000,
  enableSounds: false,
  d3SimulationEnabled: true
};

@Injectable({
  providedIn: 'root'
})
export class PreferencesService {
  private readonly platformId = inject(PLATFORM_ID);
  readonly preferences = signal<UserPreferences>(DEFAULT_PREFS);

  constructor() {
    if (isPlatformBrowser(this.platformId)) {
      this.load();
    }
  }

  update(partial: Partial<UserPreferences>): void {
    this.preferences.update(p => {
      const updated = { ...p, ...partial };
      if (isPlatformBrowser(this.platformId)) {
        localStorage.setItem('sg_preferences', JSON.stringify(updated));
      }
      return updated;
    });
  }

  private load(): void {
    try {
      const saved = localStorage.getItem('sg_preferences');
      if (saved) {
        this.preferences.set({ ...DEFAULT_PREFS, ...JSON.parse(saved) });
      }
    } catch {
      // fallback to default
    }
  }
}
