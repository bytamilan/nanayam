/**
 * @jest-environment jsdom
 */
import { getCookie, deleteCookie, orgColorClass, orgBgClass, orgBorderClass } from './auth';

describe('cookie helpers', () => {
  afterEach(() => {
    // Clear anything a test set.
    document.cookie.split(';').forEach((entry) => {
      const name = entry.split('=')[0].trim();
      if (name) deleteCookie(name);
    });
  });

  it('reads a cookie by name', () => {
    document.cookie = 'nanayam_user=alice';
    expect(getCookie('nanayam_user')).toBe('alice');
  });

  it('returns null for a cookie that is not set', () => {
    expect(getCookie('missing_cookie')).toBeNull();
  });

  it('decodes percent-encoded values', () => {
    document.cookie = 'nanayam_org=' + encodeURIComponent('ACB MSP');
    expect(getCookie('nanayam_org')).toBe('ACB MSP');
  });

  it('does not match a cookie whose name merely ends with the requested one', () => {
    document.cookie = 'other_token=wrong';
    expect(getCookie('token')).toBeNull();
  });

  it('picks the right cookie when several are set', () => {
    document.cookie = 'first=1';
    document.cookie = 'second=2';
    document.cookie = 'third=3';

    expect(getCookie('second')).toBe('2');
  });

  it('deleteCookie removes the value', () => {
    document.cookie = 'nanayam_user=alice';
    expect(getCookie('nanayam_user')).toBe('alice');

    deleteCookie('nanayam_user');

    expect(getCookie('nanayam_user')).toBeNull();
  });
});

describe('org styling helpers', () => {
  // Every organisation on the complaint channel needs a distinguishable colour,
  // otherwise the console cannot show who endorsed what.
  const orgs = ['ACBMSP', 'DeptMSP', 'OversightMSP', 'JudiciaryMSP'];

  it.each([
    ['orgColorClass', orgColorClass],
    ['orgBgClass', orgBgClass],
    ['orgBorderClass', orgBorderClass],
  ])('%s returns a distinct class per organisation', (_name, fn) => {
    const classes = orgs.map((org) => fn(org));

    expect(new Set(classes).size).toBe(orgs.length);
    classes.forEach((cls) => expect(cls).toBeTruthy());
  });

  it.each([
    ['orgColorClass', orgColorClass, 'text-slate-600'],
    ['orgBgClass', orgBgClass, 'bg-slate-600'],
    ['orgBorderClass', orgBorderClass, 'border-slate-600'],
  ])('%s falls back to a neutral class for an unknown org', (_name, fn, fallback) => {
    expect(fn('SomeFutureMSP')).toBe(fallback);
    expect(fn('')).toBe(fallback);
  });
});
