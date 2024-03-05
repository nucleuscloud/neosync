'use client';
import Checklist, { useChecklist } from '@dopt/react-checklist';

export default function OnboardingChecklist() {
  const checklist = useChecklist('newuserchecklist.three-bobcats-dress');

  console.log('checlist', checklist.items);

  return (
    <Checklist.Root>
      <Checklist.Header>
        <Checklist.Title>{checklist.title}</Checklist.Title>
        <Checklist.DismissIcon onClick={checklist.dismiss} />
      </Checklist.Header>
      <Checklist.Body>{checklist.body}</Checklist.Body>
      <Checklist.Progress
        value={checklist.count('done')}
        max={checklist.size}
      />
      <Checklist.Items>
        {checklist.items.map((item, i) => (
          <Checklist.Item key={i}>
            <Checklist.ItemIcon>
              {item.completed ? (
                <Checklist.IconCompleted />
              ) : item.skipped ? (
                <Checklist.IconSkipped />
              ) : (
                <Checklist.IconActive />
              )}
            </Checklist.ItemIcon>
            <Checklist.ItemContent>
              <Checklist.ItemTitle disabled={item.done}>
                {item.title}
              </Checklist.ItemTitle>

              <Checklist.ItemBody disabled={item.done}>
                {item.body}
              </Checklist.ItemBody>

              {!item.done && (
                <Checklist.ItemCompleteButton onClick={item.complete}>
                  {item.completeLabel}
                </Checklist.ItemCompleteButton>
              )}
            </Checklist.ItemContent>
            {!item.done && <Checklist.ItemSkipIcon onClick={item.skip} />}
          </Checklist.Item>
        ))}
      </Checklist.Items>
    </Checklist.Root>
  );
}
