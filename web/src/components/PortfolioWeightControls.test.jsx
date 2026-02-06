import { fireEvent, render, screen } from '@testing-library/react';
import { MantineProvider } from '@mantine/core';

import { PortfolioWeightControls } from './PortfolioWeightControls';

const BASELINE = {
  wavelet: 0.25,
  xgboost: 0.25,
  ridge: 0.25,
  rf: 0.25,
  svr: 0.25,
};

describe('PortfolioWeightControls', () => {
  const renderWithMantine = (ui) => render(<MantineProvider>{ui}</MantineProvider>);

  it('calls onWeightChange when slider changes', () => {
    const onWeightChange = vi.fn();
    renderWithMantine(
      <PortfolioWeightControls
        opened
        onToggle={() => {}}
        draftWeights={BASELINE}
        baselineWeights={BASELINE}
        onWeightChange={onWeightChange}
        onSave={() => {}}
        onReset={() => {}}
      />
    );

    const sliders = screen.getAllByRole('slider');
    fireEvent.keyDown(sliders[0], { key: 'ArrowRight' });
    expect(onWeightChange).toHaveBeenCalled();
  });

  it('triggers save callback and does not autosave on slider move', () => {
    const onWeightChange = vi.fn();
    const onSave = vi.fn();
    renderWithMantine(
      <PortfolioWeightControls
        opened
        onToggle={() => {}}
        draftWeights={{ ...BASELINE, wavelet: 0.4 }}
        baselineWeights={BASELINE}
        onWeightChange={onWeightChange}
        onSave={onSave}
        onReset={() => {}}
      />
    );

    const sliders = screen.getAllByRole('slider');
    fireEvent.keyDown(sliders[0], { key: 'ArrowRight' });
    expect(onSave).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole('button', { name: /save/i }));
    expect(onSave).toHaveBeenCalledTimes(1);
  });
});
