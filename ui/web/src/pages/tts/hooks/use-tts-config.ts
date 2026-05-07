// Stub for removed tts page
export interface SynthesizeParams {
  text: string;
  provider?: string;
  voice_id?: string;
  model_id?: string;
  params?: Record<string, any>;
}

export function useTtsConfig() {
  return {
    tts: {
      provider: undefined as string | undefined,
    },
    synthesize: (_params: SynthesizeParams) => Promise.resolve(new Blob())
  };
}
